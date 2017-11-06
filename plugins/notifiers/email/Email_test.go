package email

import (
	"fmt"
	"net"
	"testing"

	"github.com/gansoi/gansoi/plugins"
)

func TestNotifier(t *testing.T) {
	n := plugins.GetNotifier("email")
	_ = n.(*Email)
}

func TestNotify(t *testing.T) {
	s := NewTestServer()
	defer s.Close()

	e := &Email{
		SMTP: s.Address,
		From: "test@example.com",
		To:   "test@example.com",
	}

	err := e.Notify("test")
	if err != nil {
		t.Fatalf("Notify() failed: %s", err)
	}
}

func TestNotifyAuth(t *testing.T) {
	s := NewTestServer()
	defer s.Close()

	s.Username = "user1"
	s.Password = "password1"

	fmt.Printf("Address: %s\n", s.Address)

	e := &Email{
		SMTP:     s.Address,
		Username: "user1",
		Password: "password1",
		From:     "test@example.com",
		To:       "test@example.com",
	}

	err := e.Notify("test")
	if err != nil {
		t.Fatalf("Notify() failed: %s", err)
	}
}

func TestNotifyFail(t *testing.T) {
	s := NewTestServer()
	defer s.Close()

	cases := []Email{
		{SMTP: "127.0.0.1:82374"},
		{SMTP: s.Address},
		{SMTP: s.Address, From: "test@examaple.com"},
		{SMTP: s.Address, To: "test@examaple.com"},
		{SMTP: s.Address, From: "test@example.com", To: "test@examaple.com", Username: "blah"},
	}

	for _, e := range cases {
		err := e.Notify("test")
		if err == nil {
			t.Fatalf("Notify(%+v) did not fail", e)
		}
	}

	s.EhloResponse = "broken\r\n"
	err := (&Email{SMTP: s.Address}).Notify("test")
	if err == nil {
		t.Fatalf("Notify() did not catch broken EHLO")
	}

	s.EhloResponse = ""
	s.DataResponse = "503 Forced error\r\n"
	err = (&Email{SMTP: s.Address, From: "test@example.com", To: "test@example.com"}).Notify("test")
	if err == nil {
		t.Fatalf("Notify() did not catch broken EHLO")
	}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatalf("Listen() failed: %s", err.Error())
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		conn.Write([]byte("broken banner\r\n"))
		conn.Close()
	}()
	err = (&Email{SMTP: listener.Addr().String()}).Notify("test")
	if err == nil {
		t.Fatalf("Notify() did not catch broken banner")
	}
}

var _ plugins.Notifier = (*Email)(nil)
