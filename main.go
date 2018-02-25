package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"rsc.io/letsencrypt"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/build"
	"github.com/gansoi/gansoi/checks"
	"github.com/gansoi/gansoi/cluster"
	"github.com/gansoi/gansoi/config"
	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/eval"
	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/node"
	"github.com/gansoi/gansoi/notify"
	"github.com/gansoi/gansoi/plugins"
	_ "github.com/gansoi/gansoi/plugins/agents/error"
	_ "github.com/gansoi/gansoi/plugins/agents/filesystem"
	_ "github.com/gansoi/gansoi/plugins/agents/http"
	_ "github.com/gansoi/gansoi/plugins/agents/linuxload"
	_ "github.com/gansoi/gansoi/plugins/agents/linuxmemory"
	_ "github.com/gansoi/gansoi/plugins/agents/mysql"
	_ "github.com/gansoi/gansoi/plugins/agents/ping"
	_ "github.com/gansoi/gansoi/plugins/agents/process"
	_ "github.com/gansoi/gansoi/plugins/agents/smtp"
	_ "github.com/gansoi/gansoi/plugins/agents/ssh"
	_ "github.com/gansoi/gansoi/plugins/agents/tcpport"
	_ "github.com/gansoi/gansoi/plugins/agents/unixclock"
	_ "github.com/gansoi/gansoi/plugins/notifiers/console"
	_ "github.com/gansoi/gansoi/plugins/notifiers/email"
	_ "github.com/gansoi/gansoi/plugins/notifiers/slack"
	"github.com/gansoi/gansoi/transports/ssh"
)

var (
	configFile = config.DefaultPath

	// These can be overridden when testing.
	exit   = os.Exit
	stdin  = io.Reader(os.Stdin)
	stdout = io.Writer(os.Stdout)
	stderr = io.Writer(os.Stderr)
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
		logger.Info("main", err.Error())
		exit(1)
	}
}

func main() {
	cmdCore := &cobra.Command{
		Use:   "core",
		Short: "Run a core node",
		Long:  `Run a core node in a Gansoi cluster`,
	}
	cmdCore.PersistentFlags().StringVar(&configFile,
		"config",
		config.DefaultPath,
		"The configuration file to use.")

	coreRun := &cobra.Command{
		Use:   "run",
		Short: "Start a core node",
		Long:  "Start a core Gansoi node",
		Run:   runCore,
	}
	cmdCore.AddCommand(coreRun)

	coreInit := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new cluster",
		Long:  "Initialize a new cluster and start an internal CA",
		Run:   initCore,
	}
	cmdCore.AddCommand(coreInit)

	coreInitRun := &cobra.Command{
		Use:   "init-and-run",
		Short: "Initialize a new cluster and start it",
		Long:  "Initialize a new cluster, an internal CA and start it",
		Run:   initRunCore,
	}
	cmdCore.AddCommand(coreInitRun)

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
		Long:  "Run a Gansoi check locally and print result. This will return zero if no error occurred",
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

	cmdDemo := &cobra.Command{
		Use:   "demo",
		Short: "Run a local Gansoi demo",
		Long:  "Run a local demo accesable at https://gansoi-dev.com:9091/",
		Run:   runDemo,
	}

	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "Show version and build information",
		Run:   showVersion,
	}

	// This is added to be able to call this function without any output
	// from tests.
	hidden := &cobra.Command{
		Use:    "hidden",
		Hidden: true,
	}

	var rootCmd = &cobra.Command{Use: os.Args[0]}
	rootCmd.AddCommand(cmdCore)
	rootCmd.AddCommand(cmdCheck)
	rootCmd.AddCommand(nagCheck)
	rootCmd.AddCommand(cmdDemo)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(hidden)
	rootCmd.Execute()
}

func runCheck(printSummary bool, arguments []string) {
	var err error
	var check checks.Check
	f := stdin
	if len(arguments) > 0 {
		f, err = os.Open(arguments[0])
		if err != nil {
			stderr.Write([]byte(err.Error()))

			exit(3)
		}
	}

	in, err := ioutil.ReadAll(f)
	if err != nil {
		stderr.Write([]byte(err.Error()))

		exit(3)
	}

	// Input looks like we're called from a hash-bang script. Try to find json
	// start.
	if bytes.HasPrefix(in, []byte("#!")) {
		start := bytes.IndexRune(in, '{')
		if start == -1 {
			stderr.Write([]byte("Cannot find JSON in input stream"))

			exit(3)
		}

		in = in[start:]
	}

	err = json.Unmarshal(in, &check)
	if err != nil {
		stderr.Write([]byte(err.Error()))

		exit(3)
	}

	// FIXME: Support remote hosts here!
	result := checks.RunCheck(nil, &check)

	if printSummary {
		if result.Error != "" {
			stdout.Write([]byte(result.Error + "\n"))
		} else {
			stdout.Write([]byte("ok\n"))
		}
	}

	// Pretty-print result as json.
	out, _ := json.MarshalIndent(result, "", "\t")
	stdout.Write(out)

	if result.Error != "" {
		// Nagios and many other monitoring solutions use the exit code 2 to
		// signal check failure. Let's all agree :)
		exit(2)
	}
}

func loadConfig() *config.Configuration {
	conf := config.NewConfiguration()
	err := conf.LoadFromFile(configFile)
	if err != nil {
		logger.Info("main", "Failed to read configuration file at %s: %s", configFile, err.Error())
		exit(1)
	}

	return conf
}

func openDatabase(conf *config.Configuration) *boltdb.BoltStore {
	db, err := boltdb.NewBoltStore(path.Join(conf.DataDir, "gansoi.db"))
	if err != nil {
		logger.Info("main", "failed to open database in %s: %s", conf.DataDir, err.Error())
		exit(1)
	}

	return db
}

func initCore(_ *cobra.Command, _ []string) {
	conf := loadConfig()
	db := openDatabase(conf)
	info := cluster.NewInfo(path.Join(conf.DataDir, "cluster.json"))
	core := cluster.NewCore(info)

	self := cluster.DefaultPort(conf.Bind)
	info.SetSelf(self)

	err := core.Bootstrap()
	bailIfError(err)

	pair, err := core.Start()
	bailIfError(err)

	stream, _ := node.NewHTTPStream(conf.Bind, pair, core.CA())
	n, close, err := node.NewNode(stream, conf.DataDir, db, db, info.Self(), pair, core.CA())
	bailIfError(err)
	err = n.Bootstrap()
	bailIfError(err)
	err = close()
	bailIfError(err)
	err = db.Close()
	bailIfError(err)
}

func initRunCore(cmd *cobra.Command, arguments []string) {
	initCore(cmd, arguments)
	runCore(cmd, arguments)
}

func joinCore(_ *cobra.Command, arguments []string) {
	// Check that we have all arguments.
	if len(arguments) < 2 {
		logger.Info("join", "Too few arguments")
		exit(1)
	}

	// Split join-token in hash and cluster-token.
	parts := strings.Split(arguments[1], ".")
	if len(parts) < 2 {
		logger.Info("join", "Join-token is malformed")
		exit(1)
	}

	hash := parts[0]
	token := parts[1]

	conf := loadConfig()
	info := cluster.NewInfo(path.Join(conf.DataDir, "cluster.json"))
	core := cluster.NewCore(info)

	self := cluster.DefaultPort(conf.Bind)
	info.SetSelf(self)

	err := core.Join(arguments[0], hash, token, conf.Bind)
	bailIfError(err)
}

func printCa(_ *cobra.Command, _ []string) {
	conf := loadConfig()
	info := cluster.NewInfo(path.Join(conf.DataDir, "cluster.json"))

	stdout.Write(info.CACert)
}

func printToken(_ *cobra.Command, _ []string) {
	conf := loadConfig()
	info := cluster.NewInfo(path.Join(conf.DataDir, "cluster.json"))

	hash := sha256.Sum256(info.CACert)

	out := fmt.Sprintf("%x.%s\n", hash, info.ClusterToken)
	out += fmt.Sprintf("\nCan be used to join a new core node to the cluster like this:\n"+
		"# gansoi core join %s %x.%s\n",
		info.Self(), hash, info.ClusterToken)

	stdout.Write([]byte(out))
}

func runCore(_ *cobra.Command, _ []string) {
	conf := loadConfig()
	db := openDatabase(conf)
	info := cluster.NewInfo(path.Join(conf.DataDir, "cluster.json"))
	core := cluster.NewCore(info)

	self := cluster.DefaultPort(conf.Bind)
	info.SetSelf(self)

	pair, err := core.Start()
	bailIfError(err)

	internal := gin.New()
	internal.Use(gin.Logger())

	server := &http.Server{
		Addr: cluster.DefaultPort(conf.Bind),
		TLSConfig: &tls.Config{
			Certificates: pair,
			ClientCAs:    core.CA().CertPool(),
			ClientAuth:   tls.RequestClientCert,
		},
		Handler: internal,
	}

	go server.ListenAndServeTLS("", "")

	stream, _ := node.NewHTTPStream(conf.Bind, pair, core.CA())
	n, _, err := node.NewNode(stream, conf.DataDir, db, db, info.Self(), pair, core.CA())
	if err != nil {
		logger.Info("main", "%s", err.Error())
		exit(1)
	}
	n.Run()
	db.RegisterListener(n)

	// Do not broadcast any applies while catching up. When first booting raft
	// may choose to restore from an old snapshot and replay entries from the
	// log. Since they happened in the past, we don't want to head about them.
	db.BroadcastFrom(n.LastIndex())

	e := eval.NewEvaluator(n)

	summary := eval.NewSummary(n)
	// Seed the summarizer with all known checks.
	var all []*checks.Check
	db.All(&all, -1, 0, false)
	for _, c := range all {
		summary.AddCheck(c.ID)
	}

	// And listen for future changes.
	n.RegisterListener(summary)

	scheduler := checks.NewScheduler(n, info.Self())

	go func() {
		for {
			result := <-scheduler.Results
			e.Evaluate(result)
		}
	}()

	go func() {
		for leader := range n.LeaderCh() {
			if leader {
				sshErr := ssh.Init(n)
				if sshErr != nil {
					logger.Info("main", "ssh.Init() error: %s", sshErr.Error())
					exit(1)
				}

				conf.SaveChecks(n)
				conf.SaveHosts(n)
				conf.SaveContactGroups(n)
				conf.SaveContacts(n)

				if conf.ExclusiveSeeding {
					conf.DeleteUnknownSeeds(n)
				}

				scheduler.Run()
			} else {
				scheduler.Stop()
			}
		}
	}()

	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.ErrorLogger())

	n.Router(internal.Group("/node"))
	core.Router(internal.Group(cluster.CorePrefix), stream, n)

	api := engine.Group("/api")

	if conf.HTTP.Login != "" && conf.HTTP.Password != "" {
		api.Use(gin.BasicAuth(gin.Accounts{
			conf.HTTP.Login: conf.HTTP.Password,
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

	restHosts := node.NewRestAPI(ssh.SSH{}, n)
	restHosts.Router(api.Group("/hosts"))

	// Endpoint for running a check on the cluster node.
	api.POST("/test", func(c *gin.Context) {
		var check checks.Check
		var checkResult *checks.CheckResult
		e := c.BindJSON(&check)
		if e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
		}

		if len(check.Hosts) > 0 {
			host := ssh.SSH{}

			// Try the first host only. This should be good enough for now.
			e = db.One("ID", check.Hosts[0], &host)
			if e != nil {
				c.AbortWithError(http.StatusBadRequest, e)
				return
			}

			checkResult = checks.RunCheck(&host, &check)
			checkResult.HostID = check.Hosts[0]
		} else {
			checkResult = checks.RunCheck(nil, &check)
		}

		checkResult.Node = info.Self()

		c.JSON(http.StatusOK, checkResult)
	})

	api.POST("/testcontact", func(c *gin.Context) {
		var contact notify.Contact
		e := c.BindJSON(&contact)
		if e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
			return
		}

		e = contact.Notify(fmt.Sprintf("Test sent at %s", time.Now().String()))
		if e != nil {
			c.JSON(http.StatusBadGateway, e.Error())
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	api.POST("/testhost", func(c *gin.Context) {
		var host ssh.SSH
		e := c.BindJSON(&host)
		if e != nil {
			c.AbortWithError(http.StatusBadRequest, e)
			return
		}

		_, _, e = host.Exec("echo", "remote host access test")
		if e != nil {
			c.AbortWithError(http.StatusGatewayTimeout, e)
			return
		}

		c.JSON(http.StatusOK, nil)
	})

	api.GET("/agents", func(c *gin.Context) {
		descriptions := plugins.ListAgents()

		c.JSON(http.StatusOK, descriptions)
	})

	api.GET("/notifiers", func(c *gin.Context) {
		descriptions := plugins.ListNotifiers()

		c.JSON(http.StatusOK, descriptions)
	})

	api.GET("/backup", func(c *gin.Context) {
		t := time.Now()
		filename := fmt.Sprintf("gansoi-backup-%04d%02d%02d-%02d%02d%02d.db",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())

		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/octet-stream")

		db.WriteTo(c.Writer)
	})

	api.GET("/backup.gz", func(c *gin.Context) {
		t := time.Now()
		filename := fmt.Sprintf("gansoi-backup-%04d%02d%02d-%02d%02d%02d.db.gz",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())

		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/octet-stream")

		writer := gzip.NewWriter(c.Writer)

		db.WriteTo(writer)
		writer.Close()
	})

	engine.GET("/ssh/pubkey", func(c *gin.Context) {
		publicKey := ssh.PublicKey()

		c.Data(http.StatusOK, "text/plain", []byte(publicKey))
	})

	notifier, err := notify.NewNotifier(n)
	if err != nil {
		logger.Info("main", "Failed to start notifier: %s", err.Error())
	}
	n.RegisterListener(notifier)

	live := NewLive()
	n.RegisterListener(live)

	// Provide a websocket for clients to keep updated.
	api.GET("/live", func(c *gin.Context) {
		live.ServeHTTP(c.Writer, c.Request)
	})

	webroot := os.Getenv("GANSOI_WEBROOT")
	if webroot == "" {
		gopath := os.Getenv("GOPATH")

		webroot = gopath + "/src/github.com/gansoi/gansoi/web/"
	}

	engine.Use(static.Serve("/", static.LocalFile(webroot, true)))

	s := &http.Server{
		Addr:           conf.HTTP.Bind,
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info("main", "Binding public interface to %s", conf.HTTP.Bind)

	if conf.HTTPRedirect.Bind != "" && conf.HTTPRedirect.Target != "" {
		handler := http.RedirectHandler(conf.HTTPRedirect.Target, 307)
		go http.ListenAndServe(conf.HTTPRedirect.Bind, handler)
	}

	if conf.HTTP.TLS {
		var tlsConfig tls.Config

		if conf.HTTP.CertPath == "" || conf.HTTP.KeyPath == "" {
			var lManager letsencrypt.Manager

			cacheFile := path.Join(conf.DataDir, "letsencrypt.cache")

			if err = lManager.CacheFile(cacheFile); err != nil {
				logger.Info("main", "Failed to open Letsencrypt cachefile at %s: %s", cacheFile, err.Error())
				exit(1)
			}

			// ensure we dont ask for random certificates
			lManager.SetHosts(conf.HTTP.Hostnames)

			tlsConfig.GetCertificate = lManager.GetCertificate
		}

		s.TLSConfig = &tlsConfig

		// if GetCertificate was set earlier - ListenAndServeTLS silently ignores cert and key
		err = s.ListenAndServeTLS(conf.HTTP.CertPath, conf.HTTP.KeyPath)
	} else {
		err = s.ListenAndServe()
	}

	if err != nil {
		logger.Info("main", "Bind to %s failed: %s", conf.HTTP.Bind, err.Error())
		exit(1)
	}
}

func showVersion(_ *cobra.Command, _ []string) {
	fmt.Printf("Version:    %s\n", build.Version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Git commit: %s\n", build.ShortSHA)
	fmt.Printf("Built:      %s\n", build.Time)
}
