package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/gansoi/gansoi/cluster"
	"github.com/gansoi/gansoi/config"
)

type (
	failReader struct{}
)

func init() {
	exit = fakeExit
}

func (f *failReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error")
}

func fakeExit(code int) {
	panic(code)
}

func expectExit(t *testing.T, code int) {
	if code < 0 {
		got := recover()
		if got != nil {
			t.Fatalf("exit unexpectedly called with value %v", got)
		}

		return
	}

	got := recover()
	if got != code {
		t.Fatalf("exited with wrong error %d, expected %d", got, code)
	}
}

func TestExit(t *testing.T) {
	defer expectExit(t, 0)

	exit(0)
}

func TestBailIfError(t *testing.T) {
	defer expectExit(t, 1)

	bailIfError(errors.New("error"))
}

func TestBailIfErrorNil(t *testing.T) {
	defer expectExit(t, -1)

	bailIfError(nil)
}

func TestCobraSetup(t *testing.T) {
	os.Args = []string{os.Args[0], "hidden"}

	main()
}

func TestRunCheckJSONError(t *testing.T) {
	defer expectExit(t, 3)
	stderr = ioutil.Discard

	os.Args = []string{os.Args[0], "runcheck"}
	main()
}

func TestRunCheckMissingFile(t *testing.T) {
	defer expectExit(t, 3)
	stderr = ioutil.Discard

	os.Args = []string{os.Args[0], "runcheck", "/missing-file"}
	main()
}

func TestRunCheckReadfail(t *testing.T) {
	defer expectExit(t, 3)

	stderr = ioutil.Discard
	stdin = &failReader{}

	os.Args = []string{os.Args[0], "runcheck"}
	main()
}

func TestRunCheckBangHashJSONFail(t *testing.T) {
	defer expectExit(t, 3)

	stdout = ioutil.Discard
	stderr = ioutil.Discard
	stdin = bytes.NewBufferString(`#!/gansoi`)

	os.Args = []string{os.Args[0], "runcheck"}
	main()
}

func TestRunCheckBangHash(t *testing.T) {
	listenener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() failed: %s", err.Error())
	}
	defer listenener.Close()

	defer expectExit(t, -1)
	stdout = ioutil.Discard
	stderr = ioutil.Discard
	stdin = bytes.NewBufferString(fmt.Sprintf(`#!/gansoi
{
	"agent": "tcpport",
	"interval": 60000000000,
	"arguments": {
		"address": "%s"
	}
}`, listenener.Addr().String()))

	os.Args = []string{os.Args[0], "runcheck"}
	main()
}

func TestNagiosPluginBangHash(t *testing.T) {
	listenener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenTCP() failed: %s", err.Error())
	}
	defer listenener.Close()

	defer expectExit(t, -1)
	stdout = ioutil.Discard
	stderr = ioutil.Discard
	stdin = bytes.NewBufferString(fmt.Sprintf(`#!/gansoi
{
	"agent": "tcpport",
	"interval": 60000000000,
	"arguments": {
		"address": "%s"
	}
}`, listenener.Addr().String()))

	os.Args = []string{os.Args[0], "nagiosplugin"}
	main()
}

func TestNagiosPluginBangHashFailed(t *testing.T) {
	defer expectExit(t, 2)
	stdout = ioutil.Discard
	stderr = ioutil.Discard
	stdin = bytes.NewBufferString(`#!/gansoi
{
	"agent": "tcpport",
	"interval": 60000000000,
	"arguments": {
		"address": "127.0.0.1:0"
	}
}`)

	os.Args = []string{os.Args[0], "nagiosplugin"}
	main()
}

func TestLoadConfig(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "TestLoadConfig")
	f.WriteString("")
	f.Close()
	defer os.Remove(f.Name())

	configFile = f.Name()
	conf := loadConfig()
	if conf == nil {
		t.Errorf("loadConfig() returned nil")
	}
}

func TestLoadConfigFail(t *testing.T) {
	defer expectExit(t, 1)

	configFile = "/does/not/exist"
	loadConfig()
}

func TestOpenDatabase(t *testing.T) {
	dirname, _ := ioutil.TempDir(os.TempDir(), "TestOpenDatabase")
	defer os.RemoveAll(dirname)

	conf := &config.Configuration{
		DataDir: dirname,
	}

	store := openDatabase(conf)
	store.Close()
}

func TestOpenDatabaseFail(t *testing.T) {
	conf := &config.Configuration{
		DataDir: "/does/not/exist",
	}

	defer expectExit(t, 1)

	store := openDatabase(conf)
	store.Close()
}

func TestInitCoreFail(t *testing.T) {
	defer expectExit(t, 1)

	initCore(nil, nil)
}

func TestInitRunCoreFail(t *testing.T) {
	defer expectExit(t, 1)

	initRunCore(nil, nil)
}

func TestJoinCore(t *testing.T) {
	p1 := path.Join(os.TempDir(), "node1")
	// To test join, we must first start a working node (!)
	os.Mkdir(p1, 0700)
	defer os.RemoveAll(p1)

	c1 := `bind: "127.0.0.1:5121"
datadir: "` + p1 + `"
http:
  bind: "127.0.0.1:7121"
  tls: false
`
	// set node 1 config.
	configFile = path.Join(p1, "gansoi.conf")
	ioutil.WriteFile(configFile, []byte(c1), 0666)

	initCore(nil, nil)

	// Retrieve join-token.
	conf := loadConfig()
	info := cluster.NewInfo(path.Join(conf.DataDir, "cluster.json"))
	hash := sha256.Sum256(info.CACert)
	token := fmt.Sprintf("%x.%s", hash, info.ClusterToken)

	go runCore(nil, nil)
	time.Sleep(time.Second * 1)

	joinCore(nil, []string{"127.0.0.1:5121", token})
}

func TestJoinCoreFailNoToken(t *testing.T) {
	defer expectExit(t, 1)

	joinCore(nil, []string{"127.0.0.1:0"})
}

func TestJoinCoreFailInvalidToken(t *testing.T) {
	defer expectExit(t, 1)

	joinCore(nil, []string{"127.0.0.1:0", "no:dot"})
}

func TestPrintCAFail(t *testing.T) {
	defer expectExit(t, 1)

	printCa(nil, nil)
}

func TestMainHTTP(t *testing.T) {
	p1 := path.Join(os.TempDir(), "TestMainHTTP")
	os.Mkdir(p1, 0700)
	defer os.RemoveAll(p1)

	c1 := `bind: "127.0.0.1:5184"
datadir: "` + p1 + `"
http:
  bind: "127.0.0.1:5185"
  tls: false
`
	// set node 1 config.
	configFile = path.Join(p1, "gansoi.conf")
	ioutil.WriteFile(configFile, []byte(c1), 0644)

	initCore(nil, nil)
	go runCore(nil, nil)
	time.Sleep(time.Second * 1)

	// This could be MUCH better.
	http.Get("http://127.0.0.1:5185/api/agents")
	http.Get("http://127.0.0.1:5185/api/notifiers")
	http.Get("http://127.0.0.1:5185/api/backup")
	http.Get("http://127.0.0.1:5185/api/backup.gz")
	http.Get("http://127.0.0.1:5185/ssh/pubkey")
}

func TestMainVersion(t *testing.T) {
	defer expectExit(t, -1)

	stdout = ioutil.Discard
	stderr = ioutil.Discard

	os.Args = []string{os.Args[0], "version"}
	main()
}
