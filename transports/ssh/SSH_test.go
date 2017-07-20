package ssh

import (
	"reflect"
	"testing"
	"time"

	"github.com/gansoi/gansoi/boltdb"
	"github.com/gansoi/gansoi/transports"

	"golang.org/x/crypto/ssh"
)

func TestGenerateKey(t *testing.T) {
	key := generateKey()
	s, err := ssh.ParsePrivateKey(key)
	if err != nil {
		t.Fatalf("generateKey() did not generate a parsable key: %s", err.Error())
	}
	pub := s.PublicKey()
	if pub.Type() != "ssh-rsa" {
		t.Fatalf("Key type is wrong, format is '%s'", pub.Type())
	}
}

func TestInitFail(t *testing.T) {
	db := boltdb.NewTestStore()
	defer db.Close()

	signer = nil

	ks := keyStorage{
		ID:       "rsa-key",
		PemBytes: []byte("This is not a PEM-encoded key :-/"),
	}
	db.Save(&ks)

	err := Init(db)
	if err == nil {
		t.Fatalf("Init() did not return an error")
	}

	db = boltdb.NewTestStore()
	defer db.Close()

	db.FailSave = true
	err = Init(db)
	if err == nil {
		t.Fatalf("Init() did not return an error")
	}
}

func TestPublicKey(t *testing.T) {
	signer = nil
	if PublicKey() != "" {
		t.Fatalf("Public key is not empty: %s", PublicKey())
	}

	db := boltdb.NewTestStore()
	Init(db)

	if PublicKey() == "" {
		t.Fatalf("Public is empty")
	}
}

func TestDefaultPort(t *testing.T) {
	cases := map[string]string{
		"hello":     "hello:22",
		"hello:22":  "hello:22",
		"hello::22": "hello::22",
	}

	for input, expected := range cases {
		output := defaultPort(input)
		if output != expected {
			t.Fatalf("defaultPort() did not return what we expected, got %s, expected %s", output, expected)
		}
	}
}

func TestSSHConnect(t *testing.T) {
	serv := server{
		acceptPublicKey: true,
	}

	addr := serv.listen("127.0.0.1:0")
	defer serv.quit()

	s := &SSH{
		Address: addr,
	}

	client, err := s.connect()
	if err != nil {
		t.Fatalf("connect() returned an error: %s", err.Error())
	}

	if client == nil {
		t.Fatalf("connect() returned a nil client")
	}
}

func TestSSHConnectSlow(t *testing.T) {
	servSlow := server{
		acceptPublicKey: true,
		acceptWait:      time.Millisecond * 300,
	}
	addrSlow := servSlow.listen("127.0.0.1:0")
	defer servSlow.quit()

	servFast := server{
		acceptPublicKey: true,
	}
	addrFast := servFast.listen("127.0.0.1:0")
	defer servFast.quit()

	s1 := &SSH{
		Address: addrSlow,
	}
	go func() {
		s1.connect()
	}()
	time.Sleep(time.Millisecond * 20)

	s2 := &SSH{
		Address: addrFast,
	}

	t0 := time.Now()
	client, err := s2.connect()
	elapsed := time.Now().Sub(t0)
	if elapsed >= time.Millisecond*150 {
		t.Fatalf("fast connect (%s) appeared to be waiting for slow connect (%s), waited for %v", s2.Address, s1.Address, elapsed)
	}

	if err != nil {
		t.Fatalf("connect() returned an error: %s", err.Error())
	}

	if client == nil {
		t.Fatalf("connect() returned a nil client")
	}
}

func TestSSHConnectFail(t *testing.T) {
	s := &SSH{
		Address: "127.0.0.1:0",
	}
	client, err := s.connect()
	if err == nil {
		t.Fatalf("connect() did not return an error")
	}

	if client != nil {
		t.Fatalf("connect() returned an client when failing")
	}

	// With no signer
	signer = nil
	_, err = s.connect()
	if err == nil {
		t.Fatalf("connect() did not return an error")
	}
}

func TestSSHDial(t *testing.T) {
	_, err := (&SSH{}).Dial("tcp", "127.0.0.1:1")
	if err == nil {
		t.Fatalf("Dial did not return an error")
	}
}

func TestSSHExec(t *testing.T) {
	db := boltdb.NewTestStore()
	Init(db)

	serv := server{
		acceptPublicKey: true,
		execReply:       []byte("test"),
		execStatus:      0,
	}
	addr := serv.listen("127.0.0.1:0")
	defer serv.quit()

	s := &SSH{
		Address: addr,
	}

	_, _, err := s.Exec("echo test")
	if err != nil {
		t.Errorf("Exec() returned an unexpected error: %s", err.Error())
	}

	_, _, err = s.Exec("echo", "test")
	if err != nil {
		t.Errorf("Exec() returned an unexpected error: %s", err.Error())
	}

	if serv.connects != 1 {
		t.Errorf("Exec() connected more than once")
	}

	if serv.sessions != 2 {
		t.Errorf("Exec() started the wrong number of sessions")
	}
}

func TestExecFail(t *testing.T) {
	serv := server{
		acceptPublicKey: true,
		execReply:       []byte("test"),
		execStatus:      0,
	}
	addr := serv.listen("127.0.0.1:0")
	defer serv.quit()

	s := &SSH{
		Address: "127.0.0.1:0",
	}
	_, _, err := s.Exec("echo test")
	if err == nil {
		t.Errorf("Exec() did not catch connection error")
	}

	s.Address = addr

	serv.failChannel = true
	_, _, err = s.Exec("echo test")
	if err == nil {
		t.Errorf("Exec() did not catch channel error")
	}

	serv.failChannel = false
	serv.failExec = true
	_, _, err = s.Exec("echo test")
	if err == nil {
		t.Errorf("Exec() did not catch exec error")
	}
}

func TestSSHReadFile(t *testing.T) {
	serv := server{
		acceptPublicKey: true,
		execReply:       []byte("test output"),
		execStatus:      0,
	}
	addr := serv.listen("127.0.0.1:0")
	defer serv.quit()

	s := &SSH{
		Address: addr,
	}

	b, err := s.ReadFile("/wedontcare")
	if err != nil {
		t.Errorf("ReadFile() returned an error: %s", err.Error())
	}

	if !reflect.DeepEqual(b, serv.execReply) {
		t.Errorf("ReadFile() returned wrong contents")
	}

	serv.failExec = true
	_, err = s.ReadFile("/wedontcare")
	if err == nil {
		t.Errorf("ReadFile() did not return an error")
	}
}

var _ transports.Transport = (*SSH)(nil)
