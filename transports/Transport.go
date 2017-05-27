package transports

import (
	"io"
	"net"
)

type (
	// Transport defines a transport interface.
	Transport interface {
		// Dial should mimic net.Dial.
		Dial(network, address string) (net.Conn, error)

		// Exec should execute a binary on the host.
		Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error)

		// ReadFile should mimic ioutil.ReadFile.
		ReadFile(filename string) ([]byte, error)
	}
)
