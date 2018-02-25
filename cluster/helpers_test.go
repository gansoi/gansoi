package cluster

import (
	"testing"
)

func TestLocalIps(t *testing.T) {
	ips := localIps()

	if len(ips) < 1 {
		t.Errorf("localIps() returned an empty list")
	}
}
