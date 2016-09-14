package main

import (
	"crypto/tls"
	"flag"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/abrander/gingopherjs"
	"github.com/gin-gonic/gin"

	"rsc.io/letsencrypt"

	"github.com/abrander/gansoi/agents"
	_ "github.com/abrander/gansoi/agents/http"
	_ "github.com/abrander/gansoi/agents/tcpport"
	"github.com/abrander/gansoi/checks"
	"github.com/abrander/gansoi/database"
	"github.com/abrander/gansoi/eval"
	"github.com/abrander/gansoi/node"
)

type (
	configuration struct {
		Local       string   `toml:"local"`
		Cert        string   `toml:"cert"`
		Key         string   `toml:"key"`
		DataDir     string   `toml:"datadir"`
		Cluster     []string `toml:"cluster"`
		Secret      string   `toml:"secret"`
		LetsEncrypt bool     `toml:"letsencrypt"`
	}
)

var (
	configFile = flag.String("config", "/etc/gansoi.conf", "The configuration file to use.")

	exampleConfig = `# Example configuration for gansoi.
local = "london.example.com"
cert = "/etc/gansoi/me-cert.pem"
key = "/etc/gansoi/me-key.pem"
datadir = "/var/lib/gansoi"
cluster = ["london.example.com", "copenhagen.example.com", "berlin.example.com"]
secret = "This is unsecure. Pick a good alphanumeric secret."

# cert and key are ignored if set
letsencrypt = true
`
)

func init() {
	// This should not be used for crypto, time.Now() is enough.
	rand.Seed(time.Now().UnixNano())

	flag.Parse()
}

func main() {
	var config configuration
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		panic(err.Error())
	}

	self := config.Local

	// If local is not defined in configuration file, we use the hostname
	// according to the OS.
	if self == "" {
		self, _ = os.Hostname()
	}

	peerstore := node.NewPeerStore()
	peerstore.SetPeers(config.Cluster)
	peerstore.SetSelf(self)

	db, err := database.NewDatabase(config.DataDir)
	if err != nil {
		panic(err.Error())
	}

	n, err := node.NewNode(config.Secret, db, peerstore)
	if err != nil {
		panic(err.Error())
	}

	eval.NewEvaluator(n, peerstore)

	checks.NewScheduler(n, peerstore.Self(), true)

	engine := gin.New()

	n.Router(engine.Group("/raft"))

	restChecks := node.NewRestAPI(checks.Check{}, n)
	restChecks.Router(engine.Group("/checks"))

	restEvaluations := node.NewRestAPI(eval.Evaluation{}, n)
	restEvaluations.Router(engine.Group("/evaluations"))

	// Endpoint for running a check on the cluster node.
	engine.POST("/test", func(c *gin.Context) {
		var check checks.Check
		e := c.BindJSON(&check)
		if e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
		}

		checkResult := checks.RunCheck(&check)
		checkResult.Node = peerstore.Self()

		c.JSON(http.StatusOK, checkResult)
	})

	engine.GET("/agents", func(c *gin.Context) {
		descriptions := agents.ListAgents()

		c.JSON(http.StatusOK, descriptions)
	})

	live := NewLive()
	n.RegisterListener(live)

	// Provide a websocket for clients to keep updated.
	engine.GET("/live", func(c *gin.Context) {
		live.ServeHTTP(c.Writer, c.Request)
	})

	// This is extremely slow. Should be replaced by something else in production.
	g, _ := gingopherjs.New("github.com/abrander/gansoi/web/client")
	engine.GET("/client.js", g.Handler)

	gopath := os.Getenv("GOPATH")

	engine.StaticFile("/", gopath+"/src/github.com/abrander/gansoi/web/index.html")
	// By default we bind to port 443 (HTTPS) on all interfaecs on both IPv4
	// and IPv6.
	bind := ":443"

	// If local is defined thou, we use that instead.
	if config.Local != "" {
		bind = config.Local

		// ... and add prot 443 if needed.
		if strings.Index(bind, ":") < 0 {
			bind += ":443"
		}
	}

	var tlsConfig tls.Config

	// if letsencrypt is enabled, put a GetCertificate function into tlsConfig
	if config.LetsEncrypt {
		var lManager letsencrypt.Manager

		cacheFile := path.Join(config.DataDir, "letsencrypt.cache")

		if err := lManager.CacheFile(cacheFile); err != nil {
			panic(err.Error())
		}

		// ensure we dont ask for random certificates
		lManager.SetHosts([]string{self}) // .. or config.Local ?

		tlsConfig.GetCertificate = lManager.GetCertificate
	}

	s := &http.Server{
		Addr:           bind,
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      &tlsConfig,
	}

	// if GetCertificate was set earlier - ListenAndServeTLS silently ignores cert and key
	err = s.ListenAndServeTLS(config.Cert, config.Key)
	if err != nil {
		panic(err.Error())
	}
}
