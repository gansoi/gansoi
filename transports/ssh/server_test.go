package ssh

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

type (
	server struct {
		listener        net.Listener
		acceptPublicKey bool
		execReply       []byte
		execStatus      int
		connects        int64
		sessions        int
		failServer      bool
		failChannel     bool
		failSession     bool
		failExec        bool
		acceptWait      time.Duration
	}
)

var (
	hostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA0E8EpAVCyesVefTh1Dflu+nxF7RvtUePA3ty89i8AEXWchi0
bpION33HYestvpe1Hufv/umZv31gd3Ovptm5c6EbcujNB06gqwckC8KZQaBWPNVE
isLE/sKiHEDDpsXfIXatdCGRhd5DWI7NbejebV3e7CJ3TPDFlViJ3bSJHHuKlywH
+rbFb5i+1FEYeE6daZF3Xr6EluqBmIca+oiGfjb1KYCOBUiYrDLUggKqGHTEonOK
g/OKAqOT16dzddRUWLzMczjAabPITKH1/42yTk5I9hEIl95txcSG8Kk5ao5n+0O8
ApUXvWtrKtsTWXx5WcW0sGyCMURI2w02D0vFlQIDAQABAoIBAGkhuEHftdman/gx
M1ib7YJti5RfKJHhT7h+MYIIPLoWhSiId2fmpu4yuNIek7PBdVPn0yRgWYxfhrRC
2r/m2sYU5XmVLQUnFce7juGwkRCiD2Qcbr9plWrOaGrB9PzSmM6Wyuv1lTBvAKZN
TDjQcpHX8o5Sxss6KT46tToDx79rYEIOlqKF8Jw3sRWyCMzWvzY+VZLlLsLOy4aq
T8vVQKwEq0FlWb/l48Kaf7KI0eFEpg3T2CTTftOic0bgizcYWtsOqcCFve7W8ezr
+Y5E8HhBJDuaiFeDnmT9yhmUSsDoR/jIPdWqdhtaVdTDeyDlDKfj/3TGcc1a6gAv
nYr1fHECgYEA0QpRhBBNlnaxWgnPt36ZvcT6hZ9g8qyHhbxh3/l5eUetMOrATuJr
BSocSOQLQPU99z90RUviyvBWWhePS1DKWV5g41QDRaoGZmMrw5J+tu6IJ+I0wuQP
zkZn2IDWn7iT4rUpmlUzsw/vnG3BxOGxWE83plgoVYGf8Kbswy6d7WMCgYEA/xqf
sBXnYQwYm8N0G+ojPfekAqwuOIl5EgvWytK0fhxWIfEpuwZDH9dsr7z7veMb/z4U
TqgEFjpSX4qBdpEO6zGEJqy59wttWnOtz6ZAXw2Iy6/XaYGPm1A1xeCuQH44/TNm
tDEEDbsSOmvKVLgARg+Ynjx7Ow15IXmeAUzTjqcCgYAk8HrZKHxdc1oBvCwdk9yd
ITrX9AMQvxYvtstg5dfma5kaRNF43x/kSL24z7uBXhT2JRtpx3ArRm8r+m/S8lLf
mgLrCQSPAe2hmDK6m5+SufILgjiqm9yDKgPdyV6j2N0hObxBTn8VTGKeNPMB1JQf
M9h0p61w9rgJsaWzFADKWwKBgHhIYGURXuC2aGABX5zN4EIWouGTt9N2lnx40pUd
lncj+0TTxj5l7H51/8w5CmX1UL0f00lnFU45v59G2lZCfUtONKZbqO+gFragdqLD
G/T1l7uatLINod20izY7A05rdYmFk/aAag3jV+vt6R/bBl4Ceo6Zq+0jJCsxgdvR
EqtfAoGAIruRQ95aH+WXr2KVtV6v0itw+WlPOgPSNVW7M/KToGzZqWt8cVJ8UhF3
OBKhVLAorKhSaJKwgBL+lQtUrFAWVc+9WJi31jyor9mLl0NlNSDI6ryoyJrv4KwO
iSzceXtv5J/U8hgRd2CmnnPnR2lr/HEhA3ke+aIpLIKUqcCf17U=
-----END RSA PRIVATE KEY-----
`)
)

func (s *server) loop() {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if s.acceptPublicKey {
				return nil, nil
			}

			return nil, fmt.Errorf("denied")
		},
	}

	config.PasswordCallback = func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		return nil, errors.New("Password rejected")
	}

	key, _ := ssh.ParsePrivateKey(hostKey)
	config.AddHostKey(key)

	for {
		time.Sleep(s.acceptWait)

		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		if s.failServer {
			conn.Close()
			continue
		}
		atomic.AddInt64(&s.connects, 1)

		sshConn, channels, _, err := ssh.NewServerConn(conn, config)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}

			continue
		}
		defer sshConn.Close()

		go func() {
			for chanReq := range channels {
				go s.handleChanReq(chanReq)
			}
		}()
	}
}

func (s *server) handleChanReq(chanReq ssh.NewChannel) {
	if s.failChannel {
		chanReq.Reject(ssh.Prohibited, "failed")
		return
	}

	if chanReq.ChannelType() != "session" {
		chanReq.Reject(ssh.Prohibited, "channel type is not a session")
		return
	}

	s.sessions++

	if s.failSession {
		chanReq.Reject(ssh.Prohibited, "failed")
	}

	ch, reqs, err := chanReq.Accept()
	if err != nil {
		log.Println("fail to accept channel request", err)
		return
	}

	req := <-reqs
	if req.Type != "exec" {
		ch.Write([]byte("request type '" + req.Type + "' is not 'exec'\r\n"))
		ch.Close()
		return
	}

	s.handleExec(ch, req)
}

func (s *server) handleExec(ch ssh.Channel, req *ssh.Request) {
	if s.failExec {
		req.Reply(false, nil)
		return
	}

	if req.WantReply {
		req.Reply(true, nil)
	}

	ch.Write(s.execReply)
	ch.SendRequest("exit-status", false, ssh.Marshal(struct{ C uint32 }{C: uint32(s.execStatus)}))
	ch.Close()
}

func (s *server) listen(addr string) string {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		panic(err.Error())
	}

	go s.loop()

	return s.listener.Addr().String()
}

func (s *server) quit() {
	s.listener.Close()
}
