package main

import (
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/abrander/gansoi/agents/tcpport"
	"github.com/abrander/gansoi/database"
	"github.com/abrander/gansoi/node"
)

type (
	configuration struct {
		Local   string   `toml:"local"`
		Cert    string   `toml:"cert"`
		Key     string   `toml:"key"`
		DbPath  string   `toml:"db"`
		Cluster []string `toml:"cluster"`
		Secret  string   `toml:"secret"`
	}
)

var (
	configFile = flag.String("config", "/etc/gansoi.conf", "The configuration file to use.")

	exampleConfig = `# Example configuration for gansoi.
local = "london.example.com"
cert = "/etc/gansoi/me-cert.pem"
key = "/etc/gansoi/me-key.pem"
db = "/var/lib/gansoi"
cluster = ["london.example.com", "copenhagen.example.com", "berlin.example.com"]
secret = "This is unsecure. Pick a good alphanumeric secret."
`
)

func init() {
	flag.Parse()
}

func main() {
	var config configuration
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		panic(err.Error())
	}

	if config.Local == "" {
		panic("local not defined in confiuration file.")
	}

	peerstore := node.NewPeerStore()
	peerstore.SetPeers(config.Cluster)
	peerstore.SetSelf(config.Local)

	db, err := database.NewDatabase(config.DbPath)
	if err != nil {
		panic(err.Error())
	}

	node, err := node.NewNode(config.Secret, db, peerstore)
	if err != nil {
		panic(err.Error())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/raft", node.ServeHTTP)
	mux.HandleFunc("/raft/stats", node.Stats)
	mux.HandleFunc("/raft/apply", node.Apply)
	mux.HandleFunc("/raft/nodes", node.Nodes)

	bind := config.Local
	if strings.Index(bind, ":") < 0 {
		bind += ":443"
	}

	s := &http.Server{
		Addr:           bind,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err = s.ListenAndServeTLS(config.Cert, config.Key)
	if err != nil {
		panic(err.Error())
	}
}
