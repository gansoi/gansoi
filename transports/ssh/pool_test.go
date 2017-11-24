package ssh

import (
	"sync/atomic"
	"testing"
	"time"
)

// We need this to run before init. This is ugly.
var _ = func() (_ struct{}) {
	// For testing purposes we close connections quickly.
	closeAfter = time.Millisecond * 50
	return
}()

func TestPoolConnectFail(t *testing.T) {
	s := SSH{
		Address: "127.0.0.1:0",
	}

	client, err := connect(s)
	if err == nil {
		t.Errorf("connect() failed to detect error")
	}

	if client != nil {
		t.Errorf("connect() returned a non-nil client")
	}
}

func TestPoolConnect(t *testing.T) {
	serv := server{
		acceptPublicKey: true,
	}
	addr := serv.listen("127.0.0.1:0")
	defer serv.quit()

	s := SSH{
		Address: addr,
	}

	client, err := connect(s)
	if err != nil {
		t.Errorf("connect() returned an error: %s", err.Error())
	}

	if client == nil {
		t.Errorf("connec() returned a nil client")
	}
	client.Close()
	done(s)
}

func TestDoneFail(t *testing.T) {
	s := SSH{
		Address: "not-in-pool",
	}

	done(s)
}

func TestAutoClose(t *testing.T) {
	serv := server{
		acceptPublicKey: true,
	}
	addr := serv.listen("127.0.0.1:0")
	defer serv.quit()

	s := SSH{
		Address: addr,
	}

	client, err := connect(s)
	if err != nil {
		t.Errorf("connect() returned an error: %s", err.Error())
	}

	if client == nil {
		t.Errorf("connec() returned a nil client")
	}
	client.Close()
	done(s)

	client, err = connect(s)
	if err != nil {
		t.Errorf("connect() returned an error: %s", err.Error())
	}

	if client == nil {
		t.Errorf("connec() returned a nil client")
	}
	client.Close()
	done(s)

	if atomic.LoadInt64(&serv.connects) != 1 {
		t.Errorf("server measured %d connects, expected 1", serv.connects)
	}

	time.Sleep(time.Millisecond * 100)
	client, err = connect(s)
	if err != nil {
		t.Errorf("connect() returned an error: %s", err.Error())
	}

	if client == nil {
		t.Errorf("connec() returned a nil client")
	}
	client.Close()
	done(s)

	if serv.connects != 2 {
		t.Errorf("server measured %d connects, expected 2", serv.connects)
	}
}
