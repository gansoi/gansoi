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

// defaultPort will append the default ssh port to a hostname if needed.
func defaultPort(address string) string {
	if !strings.ContainsRune(address, ':') {
		return address + ":22"
	}

	return address
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

	start := time.Now()
	conn, err := ssh.Dial("tcp", defaultPort(s.Address), &conf)

	// This is ugly, but there's no other way of recognizing this "error".
	if err != nil && err.Error() != "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none], no supported methods remain" {
		return err
	}

	result.AddValue("HandshakeTime", ms(time.Since(start)))

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
