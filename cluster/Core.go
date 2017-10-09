package cluster

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gansoi/gansoi/ca"
	"github.com/gansoi/gansoi/logger"
	"github.com/gin-gonic/gin"
)

type (
	// Core describes a Gansoi core node.
	Core struct {
		info        *Info
		ca          *ca.CA
		pair        []tls.Certificate
		raftHandler http.Handler
		peerAdder   PeerAdder
	}
)

const (
	// CorePrefix is the http prefix core should use.
	CorePrefix = "/core"
)

// NewCore initializes a new core.
func NewCore(info *Info) *Core {
	c := &Core{
		info: info,
	}

	return c
}

// CA returns the CA.
func (c *Core) CA() *ca.CA {
	return c.ca
}

func nodeInit(info *Info, coreCA *ca.CA) ([]tls.Certificate, error) {
	// Genmerate new key for current node.
	nodeKey, err := ca.GenerateKey()
	if err != nil {
		return nil, err
	}

	// Generate CSR for current node.
	hostID := ca.RandomString(16)

	ips := localIps()
	ip := info.IP()
	if ip != nil {
		ips = append(ips, ip)
	}

	csr, err := ca.GenerateCSR(nodeKey, hostID, ips)
	if err != nil {
		return nil, err
	}

	// Sign said CSR.
	nodeCert, err := coreCA.SignCSR(csr)
	if err != nil {
		return nil, err
	}

	info.NodeCert, err = ca.EncodeCert(nodeCert)
	if err != nil {
		return nil, err
	}

	info.NodeKey, err = ca.EncodeKey(nodeKey)
	if err != nil {
		return nil, err
	}

	err = info.Save()
	if err != nil {
		return nil, err
	}

	return nodeSetup(info, coreCA)
}

func nodeSetup(info *Info, coreCA *ca.CA) ([]tls.Certificate, error) {
	var err error
	info.CAKey, err = coreCA.CertificatePEM()
	if err != nil {
		return nil, err
	}

	info.CAKey, err = coreCA.KeyPEM()
	if err != nil {
		return nil, err
	}

	pair, err := tls.X509KeyPair(info.NodeCert, info.NodeKey)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{pair}, nil
}

// Bootstrap a new cluster.
func (c *Core) Bootstrap() error {
	var err error

	if c.info.CACert != nil {
		logger.Info("main", "Cluster seem to already be initialized")
		os.Exit(1)
	}

	c.ca, err = ca.InitCA()
	if err != nil {
		return err
	}
	c.info.CACert, err = c.ca.CertificatePEM()
	if err != nil {
		return err
	}

	c.info.CAKey, err = c.ca.KeyPEM()
	if err != nil {
		return err
	}

	_, err = nodeInit(c.info, c.ca)
	if err != nil {
		return err
	}

	// Compute join token
	c.info.ClusterToken = ca.RandomString(40)
	c.info.Save()

	return nil
}

// client will return a client filled with everything we know.
func (c *Core) client() *http.Client {
	tlsConfig := &tls.Config{}

	if c.info.CACert != nil {
		// If we have establish a root, use this for verifying peer.
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(c.info.CACert)
		tlsConfig.RootCAs = pool
	} else {
		// If we don't have any root yet, use no verification.
		tlsConfig.InsecureSkipVerify = true
	}

	// If we have client certifcates, use them :)
	if c.pair != nil {
		tlsConfig.Certificates = c.pair
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{Transport: transport}
}

// Join an existing cluster.
func (c *Core) Join(address string, hash string, token string, bindPrivate string) error {
	// Get the certifcate - ignore TLS errors. We verify the cert based on the
	// hash provided in the join-token.
	resp, err := c.client().Get("https://" + DefaultPort(address) + CorePrefix + "/cert")
	if err != nil {
		return err
	}

	c.info.CACert, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if fmt.Sprintf("%x", sha256.Sum256(c.info.CACert)) != hash {
		return errors.New("Remote certifcate hash doesn't match")
	}

	_, err = ca.DecodeCert(c.info.CACert)
	if err != nil {
		return err
	}

	logger.Info("join", "Got cluster root certificate")

	// Get the root key authenticating with our cluster token.
	req, _ := http.NewRequest("GET", "https://"+DefaultPort(address)+CorePrefix+"/key", nil)
	req.Header.Add("X-Gansoi-Token", token)

	resp, err = c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c.info.CAKey, _ = ioutil.ReadAll(resp.Body)
	logger.Info("join", "Got cluster root key")

	// Set up local CA
	c.ca, err = ca.OpenCA(c.info.CAKey, c.info.CACert)
	if err != nil {
		return err
	}

	c.pair, err = nodeInit(c.info, c.ca)
	if err != nil {
		return err
	}

	logger.Info("join", "Local core initialized")

	c.info.ClusterToken = token

	// Request to join cluster.
	logger.Info("join", "Requesting raft join")
	req, _ = http.NewRequest("GET", "https://"+DefaultPort(address)+CorePrefix+"/join", nil)
	if bindPrivate != "" {
		req.Header.Add("X-Gansoi-Announce", DefaultPort(bindPrivate))
	}

	resp, err = c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(b))
	}

	logger.Info("join", "Joined.")

	return nil
}

// Start a Gansoi core node.
func (c *Core) Start() ([]tls.Certificate, error) {
	var err error

	// Set up CA.
	c.ca, err = ca.OpenCA(c.info.CAKey, c.info.CACert)
	if err != nil {
		return nil, err
	}

	return nodeSetup(c.info, c.ca)
}

// ClientCertificatePair provides a certificates used to identify this core
// node to other nodes.
func (c *Core) ClientCertificatePair() []tls.Certificate {
	// FIXME: stub
	return nil
}

func (c *Core) handleCert(context *gin.Context) {
	cert, err := c.CA().CertificatePEM()
	if err != nil {
		context.AbortWithError(500, err)
		return
	}

	context.Writer.Write(cert)
}

func (c *Core) handleKey(context *gin.Context) {
	token := context.Request.Header.Get("X-Gansoi-Token")

	if token != string(c.info.ClusterToken) {
		context.Data(401, "text/plain", []byte("token mismatch"))
		return
	}

	key, _ := c.CA().KeyPEM()
	context.Writer.Write(key)
}

func (c *Core) handleJoin(context *gin.Context) {
	name, err := c.CA().VerifyHTTPRequest(context.Request)
	if err != nil {
		context.Data(http.StatusUnauthorized, "text/plain", []byte(err.Error()))
		return
	}

	announce := context.Request.Header.Get("X-Gansoi-Announce")
	if announce == "" {
		context.Data(http.StatusBadRequest, "text/plain", []byte("Set X-Gansoi-Announce header"))
		return
	}

	logger.Debug("cluster", "%s at %s requesting to join", name, context.Request.RemoteAddr)

	err = c.peerAdder.AddPeer(announce)
	if err != nil {
		logger.Info("join", err.Error())
	}
}

func (c *Core) handleRaft(context *gin.Context) {
	name, err := c.CA().VerifyHTTPRequest(context.Request)
	if err != nil {
		context.Data(http.StatusUnauthorized, "text/plain", []byte(err.Error()))
		return
	}

	logger.Debug("internal-comm", "%s connected", name)

	c.raftHandler.ServeHTTP(context.Writer, context.Request)
}

// Router can be used to assign a Gin routergroup.
func (c *Core) Router(router *gin.RouterGroup, raftHandler http.Handler, peerAdder PeerAdder) {
	c.raftHandler = raftHandler
	c.peerAdder = peerAdder

	router.GET("/cert", c.handleCert)
	router.GET("/key", c.handleKey)
	router.GET("/join", c.handleJoin)
	router.GET("/raft", c.handleRaft)
}
