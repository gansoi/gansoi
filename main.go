package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"rsc.io/letsencrypt"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/cluster"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/eval"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/node"
	"github.com/gansoi/gansoi/notify"
	"github.com/gansoi/gansoi/plugins"
	_ "github.com/gansoi/gansoi/plugins/agents/http"
	_ "github.com/gansoi/gansoi/plugins/agents/ssh"
	_ "github.com/gansoi/gansoi/plugins/agents/tcpport"
	_ "github.com/gansoi/gansoi/plugins/notifiers/slack"
)

func init() {
	// This should not be used for crypto, time.Now() is enough.
	rand.Seed(time.Now().UnixNano())

	database.RegisterType(checks.CheckResult{})
	database.RegisterType(checks.Check{})
	database.RegisterType(notify.Contact{})
	database.RegisterType(notify.ContactGroup{})
}

func bailIfError(err error) {
	if err != nil {
		logger.Red("main", err.Error())
		os.Exit(1)
	}
}

func main() {
	cmdCore := &cobra.Command{
		Use:   "core",
		Short: "Run a core node",
		Long:  `Run a core node in a Gansoi cluster`,
		Run:   runCore,
	}
	cmdCore.PersistentFlags().StringVar(&configFile,
		"config",
		configFile,
		"The configuration file to use.")

	coreInit := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new cluster",
		Long:  "Initialize a new cluster and start an internal CA",
		Run:   initCore,
	}
	cmdCore.AddCommand(coreInit)

	corePrintCA := &cobra.Command{
		Use:   "print-ca",
		Short: "Print the CA root certificate",
		Run:   printCa,
	}
	cmdCore.AddCommand(corePrintCA)

	corePrintToken := &cobra.Command{
		Use:   "print-token",
		Short: "Print join token",
		Long: "Print the join token for this cluster node. Can be used by new\n" +
			"nodes to join this cluster.",
		Run: printToken,
	}
	cmdCore.AddCommand(corePrintToken)

	coreJoin := &cobra.Command{
		Use:   "join [leader-ip] [join-token]",
		Short: "Join a Gansoi cluster",
		Run:   joinCore,
	}
	cmdCore.AddCommand(coreJoin)

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

func loadConfig() *Configuration {
	var config Configuration
	config.SetDefaults()
	err := config.LoadFromFile(configFile)
	if err != nil {
		logger.Red("main", "Failed to read configuration file at %s: %s", configFile, err.Error())
		os.Exit(1)
	}

	return &config
}

func openDatabase(config *Configuration) *boltdb.BoltStore {
	db, err := boltdb.NewBoltStore(path.Join(config.DataDir, "gansoi.db"))
	if err != nil {
		logger.Red("main", "failed to open database in %s: %s", config.DataDir, err.Error())
		os.Exit(1)
	}

	return db
}

func initCore(cmd *cobra.Command, _ []string) {
	config := loadConfig()
	info := cluster.NewInfo(path.Join(config.DataDir, "cluster.json"))
	core := cluster.NewCore(info)

	err := core.Bootstrap()
	bailIfError(err)

	info.SetPeers([]string{cluster.DefaultPort(config.BindPrivate)})
}

func joinCore(_ *cobra.Command, arguments []string) {
	config := loadConfig()
	info := cluster.NewInfo(path.Join(config.DataDir, "cluster.json"))
	core := cluster.NewCore(info)

	// Check that we have all arguments.
	if len(arguments) < 2 {
		logger.Red("join", "Too few arguments")
		os.Exit(1)
	}

	ip := net.ParseIP(arguments[0])
	if ip == nil {
		logger.Red("join", "%s does not look like an IP address", arguments[0])
		os.Exit(1)
	}

	// Split join-token in hash and cluster-token.
	parts := strings.Split(arguments[1], ".")
	if len(parts) < 2 {
		logger.Red("join", "Join-token is malformed")
		os.Exit(1)
	}

	hash := parts[0]
	token := parts[1]

	err := core.Join(arguments[0], hash, token, config.BindPrivate)
	bailIfError(err)
}

func printCa(_ *cobra.Command, _ []string) {
	config := loadConfig()
	info := cluster.NewInfo(path.Join(config.DataDir, "cluster.json"))

	fmt.Printf("%s", info.CACert)
}

func printToken(_ *cobra.Command, _ []string) {
	config := loadConfig()
	info := cluster.NewInfo(path.Join(config.DataDir, "cluster.json"))

	hash := sha256.Sum256(info.CACert)

	fmt.Printf("%x.%s\n", hash, info.ClusterToken)

	fmt.Printf("\nCan be used to join a new core node to the cluster like this:\n"+
		"# gansoi core join %s %x.%s\n",
		info.Self(), hash, info.ClusterToken)
}

func runCore(_ *cobra.Command, _ []string) {
	config := loadConfig()
	db := openDatabase(config)
	info := cluster.NewInfo(path.Join(config.DataDir, "cluster.json"))
	core := cluster.NewCore(info)

	self := cluster.DefaultPort(config.BindPrivate)
	info.SetSelf(self)

	pair, err := core.Start()
	bailIfError(err)

	internal := gin.New()
	internal.Use(gin.Logger())

	server := &http.Server{
		Addr: cluster.DefaultPort(config.BindPrivate),
		TLSConfig: &tls.Config{
			Certificates: pair,
			ClientCAs:    core.CA().CertPool(),
			ClientAuth:   tls.RequestClientCert,
		},
		Handler: internal,
	}

	go server.ListenAndServeTLS("", "")

	stream, _ := node.NewHTTPStream(config.BindPrivate, pair, core.CA())
	n, err := node.NewNode(stream, config.DataDir, db, db, info, pair, core.CA())
	if err != nil {
		// FIXME: Fail in a more helpful manner than panic().
		panic(err.Error())
	}
	eval.NewEvaluator(n, info)

	checks.NewScheduler(n, info.Self(), true)

	engine := gin.New()
	engine.Use(gin.Logger())

	n.Router(internal.Group("/node"))
	core.Router(internal.Group(cluster.CorePrefix), stream, n)

	api := engine.Group("/api")

	if config.Login != "" && config.Password != "" {
		api.Use(gin.BasicAuth(gin.Accounts{
			config.Login: config.Password,
		}))
	}

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
		checkResult.Node = info.Self()

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

	logger.Green("main", "Binding public interface to %s", config.Bind())

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
