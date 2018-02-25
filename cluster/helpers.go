package cluster

import "net"

func localIps() []net.IP {
	var ips []net.IP

	// We can ignore errors from InterfaceAddrs. This is only used for hinting.
	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			ips = append(ips, v.IP)
		}
	}

	return ips
}
