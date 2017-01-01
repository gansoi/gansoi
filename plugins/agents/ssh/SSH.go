package ssh

import (
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/gansoi/gansoi/plugins"
)

func init() {
	plugins.RegisterAgent("ssh", SSH{})
}

// SSH will connect to a SSH server.
type SSH struct {
	Address string `json:"address" description:"The address to connect to (host or host:port)"`
}

// Check implements plugins.Agent.
func (s *SSH) Check(result plugins.AgentResult) error {
	var conf ssh.ClientConfig

	// Save host host/key info.
	conf.HostKeyCallback = func(_ string, _ net.Addr, key ssh.PublicKey) error {
		result.AddValue("KeyType", key.Type())
		result.AddValue("FingerprintMD5", ssh.FingerprintLegacyMD5(key))
		result.AddValue("FingerprintSHA256", ssh.FingerprintSHA256(key))
		return nil
	}

	address := s.Address
	if strings.IndexRune(address, ':') < 0 {
		address += ":22"
	}

	start := time.Now()
	conn, err := ssh.Dial("tcp", address, &conf)

	// This is ugly, but there's no other way of recognizing this "error".
	if err != nil && err.Error() == "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none], no supported methods remain" {
		err = nil
	}

	if err != nil {
		return err
	}

	result.AddValue("HandshakeTime", ms(time.Now().Sub(start)))

	if conn != nil {
		result.AddValue("Version", string(conn.ServerVersion()))

		return conn.Close()
	}

	return nil
}

// ms will convert a time.Duration to milliseconds.
func ms(d time.Duration) int64 {
	return ((d + time.Millisecond/2) / time.Millisecond).Nanoseconds()
}
