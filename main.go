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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"rsc.io/letsencrypt"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/eval"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/node"
	"github.com/gansoi/gansoi/notify"
	"github.com/gansoi/gansoi/plugins"
	_ "github.com/gansoi/gansoi/plugins/http"
	_ "github.com/gansoi/gansoi/plugins/notifiers/slack"
	_ "github.com/gansoi/gansoi/plugins/tcpport"
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
	var config Configuration
	config.SetDefaults()
	err := config.LoadFromFile(configFile)
	if err != nil {
		logger.Red("main", "Failed to read configuration file at %s: %s", configFile, err.Error())
		os.Exit(1)
	}

	peerstore := node.NewPeerStore()
	peerstore.SetPeers(config.Cluster)
	peerstore.SetSelf(config.Self())

	db, err := boltdb.NewBoltStore(path.Join(config.DataDir, "gansoi.db"))
	if err != nil {
		logger.Red("main", "failed to open database in %s: %s", config.DataDir, err.Error())
		os.Exit(1)
	}

	n, err := node.NewNode(config.Secret, config.DataDir, db, db, peerstore)
	if err != nil {
		// FIXME: Fail in a more helpful manner than panic().
		panic(err.Error())
	}

	eval.NewEvaluator(n, peerstore)

	checks.NewScheduler(n, peerstore.Self(), true)

	engine := gin.New()
	engine.Use(gin.Logger())

	n.Router(engine.Group("/raft"))

	api := engine.Group("/api")

	restChecks := node.NewRestAPI(checks.Check{}, n)
	restChecks.Router(api.Group("/checks"))

	restEvaluations := node.NewRestAPI(eval.Evaluation{}, n)
	restEvaluations.Router(api.Group("/evaluations"))

	restContacts := node.NewRestAPI(notify.Contact{}, n)
	restContacts.Router(api.Group("/contacts"))

	restContactGroups := node.NewRestAPI(notify.ContactGroup{}, n)
	restContactGroups.Router(api.Group("/contactgroups"))

	// Endpoint for running a check on the cluster node.
	api.POST("/test", func(c *gin.Context) {
		var check checks.Check
		e := c.BindJSON(&check)
		if e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
		}

		checkResult := checks.RunCheck(&check)
		checkResult.Node = peerstore.Self()

		c.JSON(http.StatusOK, checkResult)
	})

	api.GET("/agents", func(c *gin.Context) {
		descriptions := plugins.ListAgents()

		c.JSON(http.StatusOK, descriptions)
	})

	notifier, err := notify.NewNotifier(db)
	if err != nil {
		logger.Red("main", "Failed to start notifier: %s", err.Error())
	}
	n.RegisterClusterListener(notifier)

	live := NewLive()
	n.RegisterClusterListener(live)

	// Provide a websocket for clients to keep updated.
	api.GET("/live", func(c *gin.Context) {
		live.ServeHTTP(c.Writer, c.Request)
	})

	gopath := os.Getenv("GOPATH")

	engine.StaticFile("/", gopath+"/src/github.com/gansoi/gansoi/web/index.html")
	engine.StaticFile("/client.js", gopath+"/src/github.com/gansoi/gansoi/web/client.js")

	s := &http.Server{
		Addr:           config.Bind(),
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Green("main", "Binding to %s (Self: %s)", config.Bind(), config.Self())

	if config.TLS() {
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
			lManager.SetHosts(config.Hostnames())

			tlsConfig.GetCertificate = lManager.GetCertificate
		}

		s.TLSConfig = &tlsConfig

		// if GetCertificate was set earlier - ListenAndServeTLS silently ignores cert and key
		err = s.ListenAndServeTLS(config.Cert, config.Key)
	} else {
		err = s.ListenAndServe()
	}

	if err != nil {
		logger.Red("main", "Bind to %s failed: %s", config.Bind(), err.Error())
		os.Exit(1)
	}
}
