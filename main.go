package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/abrander/gingopherjs"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"rsc.io/letsencrypt"

	"github.com/abrander/gansoi/checks"
	"github.com/abrander/gansoi/database"
	"github.com/abrander/gansoi/eval"
	"github.com/abrander/gansoi/logger"
	"github.com/abrander/gansoi/node"
	"github.com/abrander/gansoi/notify"
	"github.com/abrander/gansoi/plugins"
	_ "github.com/abrander/gansoi/plugins/http"
	_ "github.com/abrander/gansoi/plugins/notifiers/slack"
	_ "github.com/abrander/gansoi/plugins/tcpport"
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
	configFile = "/etc/gansoi.conf"

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

	database.RegisterType(checks.CheckResult{})
	database.RegisterType(checks.Check{})
	database.RegisterType(notify.Contact{})
	database.RegisterType(notify.ContactGroup{})
}

func main() {
	cmdCore := &cobra.Command{
		Use:   "core",
		Short: "Run a core node",
		Long:  `Run a core node in a Gansoi cluster`,
		Run:   runCore,
	}
	cmdCore.Flags().StringVar(&configFile,
		"config",
		configFile,
		"The configuration file to use.")

	cmdCheck := &cobra.Command{
		Use:   "runcheck",
		Short: "Run a Gansoi check locally",
		Long:  "Run a Gansoi check locally and print result. This will return zero if no error occured",
		Run: func(_ *cobra.Command, arguments []string) {
			runCheck(false, arguments)
		},
	}

	nagCheck := &cobra.Command{
		Use:   "nagiosplugin",
		Short: "Run a Gansoi check as a nagios plugin",
		Long:  "Run a Gansoi check locally for use as a nagios plugin.",
		Run: func(_ *cobra.Command, arguments []string) {
			runCheck(true, arguments)
		},
	}

	var rootCmd = &cobra.Command{Use: os.Args[0]}
	rootCmd.AddCommand(cmdCore)
	rootCmd.AddCommand(cmdCheck)
	rootCmd.AddCommand(nagCheck)
	rootCmd.Execute()
}

func runCheck(printSummary bool, arguments []string) {
	var err error
	var check checks.Check
	f := os.Stdin
	if len(arguments) > 0 {
		f, err = os.Open(arguments[0])
		if err != nil {
			os.Stderr.WriteString(err.Error())

			os.Exit(3)
		}
	}

	in, err := ioutil.ReadAll(f)
	if err != nil {
		os.Stderr.WriteString(err.Error())

		os.Exit(3)
	}

	// Input looks like we're called from a hash-bang script. Try to find json
	// start.
	if bytes.HasPrefix(in, []byte("#!")) {
		start := bytes.IndexRune(in, '{')
		if start == -1 {
			os.Stderr.WriteString("Cannot find JSON in input stream")

			os.Exit(3)
		}

		in = in[start:]
	}

	err = json.Unmarshal(in, &check)
	if err != nil {
		os.Stderr.WriteString(err.Error())

		os.Exit(3)
	}

	result := checks.RunCheck(&check)

	if printSummary {
		if result.Error != "" {
			os.Stdout.WriteString(result.Error + "\n")
		} else {
			os.Stdout.WriteString("ok\n")
		}
	}

	// Pretty-print result as json.
	out, _ := json.MarshalIndent(result, "", "\t")
	os.Stdout.Write(out)

	if result.Error != "" {
		// Nagios and many other monitoring solutions use the exit code 2 to
		// signal check failure. Let's all agree :)
		os.Exit(2)
	}
}

func runCore(_ *cobra.Command, _ []string) {
	var config configuration
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		logger.Red("main", "Failed to read configuration file at %s: %s", configFile, err.Error())
		os.Exit(1)
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

	db, err := database.NewDatabase(path.Join(config.DataDir, "gansoi.db"))
	if err != nil {
		logger.Red("main", "failed to open database in %s: %s", config.DataDir, err.Error())
		os.Exit(1)
	}

	n, err := node.NewNode(config.Secret, config.DataDir, db, peerstore)
	if err != nil {
		// FIXME: Fail in a more helpful manner than panic().
		panic(err.Error())
	}

	eval.NewEvaluator(n, peerstore)

	checks.NewScheduler(n, peerstore.Self(), true)

	engine := gin.New()
	engine.Use(gin.Logger())

	n.Router(engine.Group("/raft"))

	restChecks := node.NewRestAPI(checks.Check{}, n)
	restChecks.Router(engine.Group("/checks"))

	restEvaluations := node.NewRestAPI(eval.Evaluation{}, n)
	restEvaluations.Router(engine.Group("/evaluations"))

	restContacts := node.NewRestAPI(notify.Contact{}, n)
	restContacts.Router(engine.Group("/contacts"))

	restContactGroups := node.NewRestAPI(notify.ContactGroup{}, n)
	restContactGroups.Router(engine.Group("/contactgroups"))

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
		descriptions := plugins.ListAgents()

		c.JSON(http.StatusOK, descriptions)
	})

	notifier, err := notify.NewNotifier(db)
	if err != nil {
		logger.Red("main", "Failed to start notifier: %s", err.Error())
	}
	n.RegisterListener(notifier)

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

		if err = lManager.CacheFile(cacheFile); err != nil {
			logger.Red("main", "Failed to open Letsencrypt cachefile at %s: %s", cacheFile, err.Error())
			os.Exit(1)
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

	logger.Green("main", "Binding to %s", bind)

	// if GetCertificate was set earlier - ListenAndServeTLS silently ignores cert and key
	err = s.ListenAndServeTLS(config.Cert, config.Key)
	if err != nil {
		logger.Red("main", "Bind to %s failed: %s", bind, err.Error())
		os.Exit(1)
	}
}
