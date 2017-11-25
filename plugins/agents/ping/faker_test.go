package ping

import (
	"errors"
	"net"
	"strings"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type (
	faker struct {
		network  string
		incoming chan []byte
		outgoing chan []byte
		closed   chan struct{}
	}
)

func newFaker(network string) *faker {
	f := &faker{
		network:  network,
		incoming: make(chan []byte, 100),
		outgoing: make(chan []byte, 100),
		closed:   make(chan struct{}, 1),
	}

	go f.loop()

	return f
}

func newListener(listenError4 error, listenError6 error) func(string, string) (readwritecloser, error) {
	return func(network, address string) (readwritecloser, error) {
		if listenError4 != nil && strings.HasPrefix(network, "ip4") {
			return nil, listenError4
		}

		if listenError6 != nil && strings.HasPrefix(network, "ip6") {
			return nil, listenError6
		}

		return newFaker(network), nil
	}
}

func (f *faker) ReadFrom(b []byte) (int, net.Addr, error) {
	select {
	case in := <-f.outgoing:
		copy(b, in)
		return len(in), &net.IPAddr{}, nil

	case <-f.closed:
		switch {
		case strings.HasPrefix(f.network, "ip4"):
			return 0, nil, errors.New("read ip4 0.0.0.0: use of closed network connection")
		case strings.HasPrefix(f.network, "ip6"):
			return 0, nil, errors.New("read ip6 ::: use of closed network connection")
		default:
			return 0, nil, errors.New("closed")
		}
	}

}

func (f *faker) WriteTo(b []byte, dst net.Addr) (int, error) {
	if dst.String() == "127.0.0.2" {
		return len(b), nil
	}

	f.incoming <- b

	m4, _ := icmp.ParseMessage(1, b)
	m6, _ := icmp.ParseMessage(58, b)

	body, ok := m4.Body.(*icmp.Echo)
	if ok {
		reply, _ := (&icmp.Message{
			Type: ipv4.ICMPTypeEchoReply,
			Code: 0,
			Body: &icmp.Echo{
				ID:   body.ID,
				Seq:  body.Seq,
				Data: body.Data,
			},
		}).Marshal(nil)

		f.outgoing <- reply
	}

	body, ok = m6.Body.(*icmp.Echo)
	if ok {
		// FIXME: ipv6
	}

	return len(b), nil
}

func (f *faker) Close() error {
	close(f.incoming)

	f.closed <- struct{}{}

	return nil
}

func (f *faker) loop() {
	for pkg := range f.incoming {
		f.outgoing <- pkg
	}
}
