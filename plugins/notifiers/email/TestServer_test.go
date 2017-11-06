package email

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"testing"
)

type (
	TestServer struct {
		Address       string
		Username      string
		Password      string
		EhloResponse  string
		DataResponse  string
		listener      net.Listener
		authenticated bool
	}
)

func NewTestServer() *TestServer {
	t := &TestServer{}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err.Error())
	}

	t.Address = listener.Addr().String()
	t.listener = listener

	go t.run()

	return t
}

func (t *TestServer) run() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			continue
		}

		buf := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		buf.WriteString("220 test-server ESMTP\r\n")
		buf.Flush()

		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				break
			}

			fields := strings.Fields(line)
			cmd := strings.ToUpper(fields[0])

			switch cmd {
			case "HELO":
				t.HELO(fields, buf)
			case "EHLO":
				t.EHLO(fields, buf)
			case "AUTH":
				t.AUTH(fields, buf)
			case "MAIL", "RCPT":
				t.MAIL(fields, buf, conn)
			case "DATA":
				t.DATA(fields, buf)
			case "QUIT":
				buf.WriteString("221 Good luck.\r\n")
				buf.Flush()
				conn.Close()
			default:
				buf.WriteString("502 Command not recognized\r\n")
				buf.Flush()
			}
		}
	}
}

func (t *TestServer) HELO(parts []string, buf *bufio.ReadWriter) {
	if t.EhloResponse != "" {
		buf.WriteString(t.EhloResponse)
		buf.Flush()
		return
	}

	buf.WriteString("250 test-server\r\n")
	buf.Flush()
}

func (t *TestServer) EHLO(fields []string, buf *bufio.ReadWriter) {
	if t.EhloResponse != "" {
		buf.WriteString(t.EhloResponse)
		buf.Flush()
		return
	}

	buf.WriteString("250-test-server\r\n")
	buf.WriteString("250-AUTH PLAIN LOGIN\r\n")
	buf.WriteString("250-AUTH=PLAIN LOGIN\r\n")
	buf.WriteString("250 8BITMIME\r\n")
	buf.Flush()
}

func (t *TestServer) AUTH(fields []string, w *bufio.ReadWriter) {
	s, _ := base64.StdEncoding.DecodeString(fields[2])
	truth := fmt.Sprintf("\000%s\000%s", t.Username, t.Password)

	if string(s) != truth {
		response := fmt.Sprintf("535 Not authenticated... (%s != %s)\r\n", s, truth)
		w.WriteString(response)
		w.Flush()
		return
	}

	t.authenticated = true
	w.WriteString("235 Ok. Authenticated\r\n")
	w.Flush()
}

func (t *TestServer) MAIL(fields []string, buf *bufio.ReadWriter, conn net.Conn) {
	if t.Username != "" && !t.authenticated {
		buf.WriteString("454 Denied. Please authenticate\r\n")
		buf.Flush()
		conn.Close()
	}

	if strings.Contains(fields[1], "<>") {
		buf.WriteString("501 no address\r\n")
		buf.Flush()
		return
	}

	buf.WriteString("250 ok\r\n")
	buf.Flush()
}

func (t *TestServer) DATA(_ []string, buf *bufio.ReadWriter) {
	if t.DataResponse != "" {
		buf.WriteString(t.DataResponse)
		buf.Flush()
		return
	}
	buf.WriteString("354 ok. Go ahead. End with <CR><LF>.<CR><LF>\r\n")
	buf.Flush()
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}

		if line == ".\r\n" {
			break
		}
	}

	buf.WriteString("250 ok. Got it.\r\n")
	buf.Flush()
}

func (t *TestServer) Close() {
	t.listener.Close()
}

func TestTestServer(t *testing.T) {
	s := NewTestServer()

	err := smtp.SendMail(s.Address, nil, "test@example.com", []string{"test@example.com"}, []byte("Hello there"))
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
}
