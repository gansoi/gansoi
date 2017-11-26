package ping

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"golang.org/x/net/icmp"
)

type (
	mock struct {
		ret     error
		packets [][]byte
	}
)

func (m *mock) WriteTo(b []byte, dst net.Addr) (int, error) {
	m.packets = append(m.packets, b)

	return len(b), m.ret
}

func TestLookup4(t *testing.T) {
	ipv4, ipv6, err := lookup("go-test-target-v4.gansoi-dev.com")
	if err != nil {
		t.Fatalf("lookup() failed: %s", err.Error())
	}

	if len(ipv4) != 1 {
		t.Fatalf("lookip() returned wrong number of IPv4 addresses (%d) expected 1: %+v", len(ipv4), ipv4)
	}

	if len(ipv6) != 0 {
		t.Fatalf("lookip() returned wrong number of IPv6 addresses (%d) expected 0: %+v", len(ipv6), ipv6)
	}

	if ipv4[0].String() != "198.51.100.1" {
		t.Fatalf("lookup() returned wrong ipv4 address. Expected 198.51.100.1, got %s", ipv4[0].String())
	}
}

func TestLookup6(t *testing.T) {
	ipv4, ipv6, err := lookup("go-test-target-v6.gansoi-dev.com")
	if err != nil {
		t.Fatalf("lookup() failed: %s", err.Error())
	}

	if len(ipv4) != 0 {
		t.Fatalf("lookip() returned wrong number of IPv4 addresses (%d) expected 0: %+v", len(ipv4), ipv4)
	}

	if len(ipv6) != 1 {
		t.Fatalf("lookip() returned wrong number of IPv6 addresses (%d) expected 1: %+v", len(ipv6), ipv6)
	}

	if ipv6[0].String() != "2001:db8::2" {
		t.Fatalf("lookup() returned wrong ipv4 address. Expected 2001:db8::2, got %s", ipv6[0].String())
	}
}

func TestLookup46(t *testing.T) {
	ipv4, ipv6, err := lookup("go-test-target.gansoi-dev.com")
	if err != nil {
		t.Fatalf("lookup() failed: %s", err.Error())
	}

	if len(ipv4) != 1 {
		t.Fatalf("lookip() returned wrong number of IPv4 addresses (%d) expected 1: %+v", len(ipv4), ipv4)
	}

	if len(ipv6) != 1 {
		t.Fatalf("lookip() returned wrong number of IPv6 addresses (%d) expected 1: %+v", len(ipv6), ipv6)
	}

	if ipv4[0].String() != "198.51.100.1" {
		t.Fatalf("lookup() returned wrong ipv4 address. Expected 198.51.100.1, got %s", ipv4[0].String())
	}
}

func TestLookupIP4(t *testing.T) {
	ipv4, ipv6, err := lookup("198.51.100.51")
	if err != nil {
		t.Fatalf("lookup() failed: %s", err.Error())
	}

	if len(ipv4) != 1 {
		t.Fatalf("lookip() returned wrong number of IPv4 addresses (%d) expected 1: %+v", len(ipv4), ipv4)
	}

	if len(ipv6) != 0 {
		t.Fatalf("lookip() returned wrong number of IPv6 addresses (%d) expected 0: %+v", len(ipv6), ipv6)
	}

	if ipv4[0].String() != "198.51.100.51" {
		t.Fatalf("lookup() returned wrong ipv4 address. Expected 198.51.100.51, got %s", ipv4[0].String())
	}
}

func TestLookupIP6(t *testing.T) {
	ipv4, ipv6, err := lookup("2001:db8::52")
	if err != nil {
		t.Fatalf("lookup() failed: %s", err.Error())
	}

	if len(ipv4) != 0 {
		t.Fatalf("lookip() returned wrong number of IPv4 addresses (%d) expected 0: %+v", len(ipv4), ipv4)
	}

	if len(ipv6) != 1 {
		t.Fatalf("lookip() returned wrong number of IPv6 addresses (%d) expected 1: %+v", len(ipv6), ipv6)
	}

	if ipv6[0].String() != "2001:db8::52" {
		t.Fatalf("lookup() returned wrong ipv4 address. Expected 2001:db8::52, got %s", ipv6[0].String())
	}
}

func TestLookupFail(t *testing.T) {
	ipv4, ipv6, err := lookup("go-test-nonexisting.gansoi-dev.com")
	if err == nil {
		t.Fatalf("lookup() did not err on nonexisting hostname")
	}

	if ipv4 != nil {
		t.Fatalf("lookup() did not return empty ipv4 slice for nonexisting hostname")
	}

	if ipv6 != nil {
		t.Fatalf("lookup() did not return empty ipv6 slice for nonexisting hostname")
	}
}

func TestICMPServiceNextID(t *testing.T) {
	previousID = 0

	for i := 1; i < 100000; i++ {
		got := nextID()
		if got != uint16(i&0xffff) {
			t.Fatalf("nextID() returned something unexpected, expected %d, got %d", i, got)
		}
	}
}

func TestICMPServiceStartStop(t *testing.T) {
	a := available
	available = true
	defer func() { available = a }()

	listenPacket = newListener(nil, nil)
	defer func() { listenPacket = listen }()

	i := NewICMPService()
	err := i.Start()
	if err != nil {
		t.Fatalf("Start() failed: %s", err.Error())
	}

	time.Sleep(time.Millisecond * 100)

	i.Stop()

	i.Start()
	i.conn4.(*faker).network = "buh"
	i.Stop()
}

func TestNewICMPPacket4(t *testing.T) {
	p := newICMPPacket4(50, 1500)

	packet := gopacket.NewPacket(p, layers.LayerTypeICMPv4, gopacket.NoCopy)

	l := packet.Layers()

	if len(l) != 2 {
		t.Fatalf("Wrong number of layers in packet. Expected 2, got %d", len(l))
	}

	if l[0].LayerType() != layers.LayerTypeICMPv4 {
		t.Fatalf("LayerType is not layers.LayerTypeICMPv4, got %s", l[0].LayerType().String())
	}

	icmp := l[0].(*layers.ICMPv4)
	if icmp.Id != 50 {
		t.Fatalf("Wrong ID encoded. Got %d, expected 50", icmp.Id)
	}

	if icmp.Seq != 1500 {
		t.Fatalf("Wrong ID encoded. Got %d, expected 1500", icmp.Id)
	}

	payload := ICMPPayload{}
	err := payload.Read(icmp.Payload)
	if err != nil {
		t.Fatalf("newICMPPacket4() failed to encode proper payload: %s", err.Error())
	}
}

func TestNewICMPPacket6(t *testing.T) {
	p := newICMPPacket6(50, 1500)

	packet := gopacket.NewPacket(p, layers.LayerTypeICMPv6, gopacket.NoCopy)

	l := packet.Layers()

	if len(l) != 2 {
		t.Fatalf("Wrong number of layers in packet. Expected 2, got %d", len(l))
	}

	if l[0].LayerType() != layers.LayerTypeICMPv6 {
		t.Fatalf("LayerType is not layers.LayerTypeICMPv6, got %s", l[0].LayerType().String())
	}

	_, err := icmp.ParseMessage(58, p)
	if err != nil {
		t.Fatalf("Failed to parse packet: %s", err.Error())
	}
}

func TestSendEchoRequest4(t *testing.T) {
	m := &mock{
		ret: nil,
	}

	err := sendEchoRequest4(m, 150, 120, nil)

	if err != nil {
		t.Fatalf("sendEchoRequest4() failed: %s", err.Error())
	}

	if len(m.packets) != 120 {
		t.Fatalf("Wrong number of packets transmitted, got %d, expected 120", len(m.packets))
	}

	// TODO: Check the content of the packages at some point.

	m.ret = errors.New("mock error")

	err = sendEchoRequest4(m, 150, 120, nil)

	if err == nil {
		t.Fatalf("sendEchoRequest4() failed to catch transport error")
	}
}

func TestSendEchoRequest6(t *testing.T) {
	m := &mock{
		ret: nil,
	}

	err := sendEchoRequest6(m, 150, 120, nil)

	if err != nil {
		t.Fatalf("sendEchoRequest6() failed: %s", err.Error())
	}

	if len(m.packets) != 120 {
		t.Fatalf("Wrong number of packets transmitted, got %d, expected 120", len(m.packets))
	}

	// TODO: Check the content of the packages at some point.

	m.ret = errors.New("mock error")

	err = sendEchoRequest6(m, 150, 120, nil)

	if err == nil {
		t.Fatalf("sendEchoRequest6() failed to catch transport error")
	}
}

func TestAvailable(t *testing.T) {
	listenPacket = newListener(nil, nil)
	defer func() { listenPacket = listen }()
	if !Available() {
		t.Fatalf("Available() returned false, expected true")
	}

	listenPacket = newListener(ErrICMPServiceUnavailable, nil)
	if Available() {
		t.Fatalf("Available() returned true, expected false")
	}
}

func TestStartFail(t *testing.T) {
	a := available
	available = true
	defer func() { available = a }()

	err := errors.New("listen ip4:icmp 0.0.0.0: socket: operation not permitted")
	listenPacket = newListener(err, nil)
	defer func() { listenPacket = listen }()

	i := NewICMPService()
	i.Start()

	err = errors.New("error")
	listenPacket = newListener(err, nil)
	i.Start()

	err = errors.New("error")
	listenPacket = newListener(nil, err)
	i.Start()
}

func TestPingAvailability(t *testing.T) {
	a := available
	i := NewICMPService()
	available = false

	_, err := i.Ping("127.0.0.1", 1, 0)
	if err != ErrICMPServiceUnavailable {
		t.Errorf("Ping() failed to return correct error when ICMP unavailable")
	}

	available = a
}

func TestGotReply(t *testing.T) {
	p := NewICMPPayload()
	i := NewICMPService()
	i.active[123] = make(chan *icmpReply, 2)
	i.gotReply(123, []byte{})
	i.gotReply(123, p.Bytes())
}

func TestPingOverflow(t *testing.T) {
	a := available
	available = true
	defer func() { available = a }()

	prev := atomic.LoadInt32(&previousID)

	i := NewICMPService()
	i.Start()

	i.active[uint16(prev+1)] = nil
	_, err := i.Ping("127.0.0.1", 1, time.Millisecond)
	if err == nil {
		t.Errorf("Failed to detect ICMP ID overflow")
	}
	delete(i.active, uint16(prev+1))
}
