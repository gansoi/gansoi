package unixclock

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gansoi/gansoi/plugins"
	"github.com/gansoi/gansoi/transports"
)

type (
	// UnixClock checks for clock drift on unix hosts.
	UnixClock struct{}
)

func init() {
	plugins.RegisterAgent("unixclock", UnixClock{})
}

// RemoteCheck implements RemoteAgent.
func (l *UnixClock) RemoteCheck(transport transports.Transport, result plugins.AgentResult) error {
	out, _, err := transport.Exec("date", "+%s")
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(out)
	if err != nil {
		return err
	}

	trimmed := strings.TrimSpace(string(b))
	remoteTime, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return err
	}

	skew := remoteTime - time.Now().Unix()
	if skew < 0 {
		skew *= -1
	}

	result.AddValue("ClockSkew", int(skew))

	return nil
}
