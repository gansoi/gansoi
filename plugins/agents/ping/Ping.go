package ping

import (
	"time"

	"github.com/gansoi/gansoi/logger"
	"github.com/gansoi/gansoi/plugins"
)

func init() {
	if Available() {
		plugins.RegisterAgent("ping", Ping{})

		i = NewICMPService()
		i.Start()
	} else {
		logger.Info("ping", "ICMP ping not available")
	}
}

type (
	// Ping will try to "ping" the host using ICMP echo.
	Ping struct {
		Target string `json:"target" description:"Target to ping"`
		Count  int    `json:"count" description:"Number of ICMP echo packets to send"`
	}
)

var (
	i *ICMPService
)

// Check implements plugins.Agent.
func (p *Ping) Check(result plugins.AgentResult) error {
	status, err := i.Ping(p.Target, p.Count, time.Second*5)
	if err != nil {
		return err
	}

	result.AddValue("Min", ms(status.Min))
	result.AddValue("Max", ms(status.Max))
	result.AddValue("Average", ms(status.Average))
	result.AddValue("PacketLoss", 1-float64(status.Replies)/float64(status.Sent))
	result.AddValue("Sent", status.Sent)
	result.AddValue("Replies", status.Replies)

	return nil
}

func ms(t time.Duration) float64 {
	return float64(t) / float64(time.Millisecond)
}
