package cluster

import "net"

func localIps() []net.IP {
	var ips []net.IP

	// We can ignore errors from InterfaceAddrs. This is only used for hinting.
	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		if v, ok := addr.(*net.IPNet); ok {
			ips = append(ips, v.IP)
		}
	}

	return ips
}
