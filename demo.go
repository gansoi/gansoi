package main

import (
	"io/ioutil"
	"os"
	"path"
	"text/template"

	_ "embed"

	"github.com/gansoi/gansoi/logger"
	"github.com/spf13/cobra"
)

//go:embed dockerroot/gansoi-dev.com-cert.pem
var demoCert []byte

//go:embed dockerroot/gansoi-dev.com-key.pem
var demoKey []byte

var demoConfig = `
bind: "127.0.0.1:0"
datadir: "{{ .Datadir }}"

http:
  bind: "gansoi-dev.com:9002"
  hostnames:
    - "gansoi-dev.com"
  cert: "{{ .Datadir }}/cert.pem"
  key: "{{ .Datadir }}/key.pem"

checks:
  gansoi.com-https:
    agent: "http"
    arguments:
      url: "https://gansoi.com/"
    expressions:
      - "StatusCode == 200"
      - "TimeAccumulated < 1500"

  gansoi-com-http:
    agent: "http"
    arguments:
      url: "http://gansoi.com/"
      followRedirect: false
    expressions:
      - "StatusCode == 302"
      - "TimeAccumulated < 1500"

  gansoi-com-ssh:
    agent: "ssh"
    arguments:
      address: "gansoi.com"
    expressions:
      - "HandshakeTime < 1000"
`

func runDemo(cmd *cobra.Command, arguments []string) {
	cwd, _ := os.Getwd()

	dataDir := path.Join(cwd, "demo-data")
	configFile = path.Join(dataDir, "config.yml")

	os.Mkdir(dataDir, 0755)

	ioutil.WriteFile(path.Join(dataDir, "cert.pem"), demoCert, 0644)
	ioutil.WriteFile(path.Join(dataDir, "key.pem"), demoKey, 0644)

	wr, _ := os.OpenFile(configFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	data := map[string]interface{}{
		"Datadir": dataDir,
	}

	err := template.Must(template.New("name").Parse(demoConfig)).Execute(wr, data)
	if err != nil {
		logger.Info("demo", "Failed to write demo configuration: %s", err.Error())
		exit(1)
	}

	_, err = os.Stat(path.Join(dataDir, "cluster.json"))
	if err != nil {
		initCore(cmd, arguments)
	}

	runCore(cmd, arguments)
}
