package ping

import (
	"errors"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/gansoi/gansoi/logger"
)

type (
	// ICMPService is a service capable of pinging remotes and listening for
	// answers.
	ICMPService struct {
		conn4      readwritecloser
		conn6      readwritecloser
		activeLock sync.RWMutex
		active     map[uint16]chan *icmpReply
	}

	icmpReply struct {
		RTT time.Duration
	}

	// ICMPSummary will be returned from Ping() to give a quick summary.
	ICMPSummary struct {
		Sent    int
		Replies int
		Min     time.Duration
		Max     time.Duration
		Average time.Duration
	}

	// We create a few interfaces to make testing possible.
	reader interface {
		ReadFrom(b []byte) (int, net.Addr, error)
	}

	writer interface {
		WriteTo(b []byte, dst net.Addr) (int, error)
	}

	closer interface {
		Close() error
	}

	readwritecloser interface {
		reader
		writer
		closer
	}
)

var (
	// No matter how many ICMPService's that is started, we must have unique
	// ID's across all instances. That's why this is a package variable.
	previousID int32

	// ErrICMPServiceUnavailable will be returned if the process doesn't have
	// sufficient privileges.
	ErrICMPServiceUnavailable = errors.New("ICMP service unavailable")

	// This will be set as true in init() if ICMP is allowed.
	available bool
)

func init() {
	available = Available()
}

// Available returns true if the ICMP service is available.
func Available() bool {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err == nil && conn != nil {
		conn.Close()

		return true
	}

	return false
}

func nextID() uint16 {
	// Return the lower 16 bit as id.
	return uint16(atomic.AddInt32(&previousID, 1) & 0xffff)
}

// lookup looks up a hostname using the system resolver.
func lookup(addr string) (ipv4 []net.Addr, ipv6 []net.Addr, err error) {
	hits, err := net.LookupHost(addr)
	if err != nil {
		return nil, nil, err
	}

	for _, hit := range hits {
		ip, _ := net.ResolveIPAddr("ip", hit)
		// According to the documentation To4() returns nil if the IP address
		// cannot be expressed as IPv4.
		switch ip.IP.To4() {
		case nil:
			ipv6 = append(ipv6, ip)
		default:
			ipv4 = append(ipv4, ip)
		}
	}

	return ipv4, ipv6, nil
}

// NewICMPService instantiates a new ICMPService.
func NewICMPService() *ICMPService {
	i := &ICMPService{
		active: make(map[uint16]chan *icmpReply),
	}

	return i
}

// Start starts the ICMPService.
func (i *ICMPService) Start() {
	var err error
	i.conn4, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		if err.Error() == "listen ip4:icmp 0.0.0.0: socket: operation not permitted" {
			logger.Info("icmpping", "Please run:\nsudo setcap cap_net_raw=ep %s\n", os.Args[0])
		} else {
			logger.Info("icmpping", err.Error())
		}
	}

	go listenLoop(i.conn4, i.processPacket4)

	i.conn6, err = icmp.ListenPacket("ip6:ipv6-icmp", "")
	if err != nil {
		logger.Info("ping", "IPv6 ICMP seem unavailable: %s", err.Error())
	}

	go listenLoop(i.conn6, i.processPacket6)
}

func (i *ICMPService) gotReply(id uint16, payload []byte) {
	i.activeLock.RLock()
	ch, found := i.active[id]
	i.activeLock.RUnlock()

	if found {
		p := &ICMPPayload{}
		err := p.Read(payload)
		if err != nil {
			logger.Debug("ping", "Payload error: %s\n", err.Error())
			return
		}

		ch <- &icmpReply{RTT: time.Since(p.Timestamp)}
		return
	}
}

func (i *ICMPService) processPacket4(bytes []byte) {
	packet := gopacket.NewPacket(bytes, layers.LayerTypeICMPv4, gopacket.NoCopy)

	ll := packet.Layers()

	if len(ll) > 0 && ll[0].LayerType() == layers.LayerTypeICMPv4 {
		icmp := ll[0].(*layers.ICMPv4)
		typ := uint8(icmp.TypeCode >> 8)

		if typ == layers.ICMPv4TypeEchoReply {
			i.gotReply(icmp.Id, icmp.Payload)
		}
	}
}

func (i *ICMPService) processPacket6(bytes []byte) {
	// Protocol 58 is IPv6-ICMP as described in rfc 2460.
	m, _ := icmp.ParseMessage(58, bytes)

	if m.Type != ipv6.ICMPTypeEchoReply {
		return
	}

	switch packet := m.Body.(type) {
	case *icmp.Echo:
		i.gotReply(uint16(packet.ID), packet.Data)

	default:
		logger.Info("ping", "Got %T, doing nothing", packet)
		return
	}
}

func listenLoop(conn reader, processPacket func([]byte)) {
	// Set maximum packet size to 9000 to support jumbo frames
	readBytes := make([]byte, 9000)

	for {
		n, _, err := conn.ReadFrom(readBytes)
		// Take care of IPv4 closed connection. This is ugly.
		if err != nil && err.Error() == "read ip4 0.0.0.0: use of closed network connection" {
			break
		}

		// Take care of IPv6 closed connection.
		if err != nil && err.Error() == "read ip6 ::: use of closed network connection" {
			break
		}

		if err != nil {
			logger.Info("ping", "listenLoop, error from ReadFrom(): %s", err.Error())

			time.Sleep(time.Millisecond * 100)
			continue
		}

		processPacket(readBytes[:n])
	}
}

// Stop stops the listening loop.
func (i *ICMPService) Stop() error {
	err := i.conn4.Close()
	if err != nil {
		return err
	}

	return i.conn6.Close()
}

func newICMPPacket4(id uint16, seq int) []byte {
	b, _ := (&icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(id),
			Seq:  seq,
			Data: NewICMPPayload().Bytes(),
		},
	}).Marshal(nil)

	return b
}

func sendEchoRequest4(conn writer, id uint16, count int, target net.Addr) error {
	for seq := 0; seq < count; seq++ {
		b := newICMPPacket4(id, seq)

		_, err := conn.WriteTo(b, target)
		if err != nil {
			return err
		}
	}

	return nil
}

func newICMPPacket6(id uint16, seq int) []byte {
	b, _ := (&icmp.Message{
		Type: ipv6.ICMPTypeEchoRequest,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(id),
			Seq:  seq,
			Data: NewICMPPayload().Bytes(),
		},
	}).Marshal(nil)

	return b
}

func sendEchoRequest6(conn writer, id uint16, count int, target net.Addr) error {
	for seq := 0; seq < count; seq++ {
		b := newICMPPacket6(id, seq)

		_, err := conn.WriteTo(b, target)
		if err != nil {
			return err
		}
	}

	return nil
}

// Ping will ping the target using ICMP echo/reply.
func (i *ICMPService) Ping(target string, count int, timeout time.Duration) (*ICMPSummary, error) {
	if !available {
		return nil, ErrICMPServiceUnavailable
	}

	targets4, targets6, err := lookup(target)
	if err != nil {
		return nil, err
	}

	status := &ICMPSummary{}
	status.Sent = (len(targets4) + len(targets6)) * count
	replyChannel := make(chan *icmpReply, status.Sent)

	id := nextID()

	i.activeLock.Lock()
	_, found := i.active[id]
	if found {
		i.activeLock.Unlock()
		return nil, errors.New("ICMP ID overflow")
	}
	i.active[id] = replyChannel
	i.activeLock.Unlock()

	for _, target := range targets4 {
		sendEchoRequest4(i.conn4, id, count, target)
	}

	for _, target := range targets6 {
		sendEchoRequest6(i.conn6, id, count, target)
	}

	var rtts []time.Duration

	t := time.After(timeout)
OUTER:
	for {
		select {
		case reply := <-replyChannel:
			rtts = append(rtts, reply.RTT)

			status.Replies++
			if status.Replies == status.Sent {
				break OUTER
			}
		case <-t:
			break OUTER
		}
	}

	i.activeLock.Lock()
	delete(i.active, id)
	i.activeLock.Unlock()

	// Compute min, average and max
	status.Min = time.Hour
	var sum time.Duration
	for _, rtt := range rtts {
		if rtt < status.Min {
			status.Min = rtt
		}

		if rtt > status.Max {
			status.Max = rtt
		}

		sum += rtt
	}

	// Average only makes sense if there's any replies.
	if status.Replies > 0 {
		status.Average = sum / time.Duration(status.Replies)
	}

	return status, nil
}
