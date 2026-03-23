package netutil

import (
	"log"
	"net"
)

var (
	dialFunc      = net.Dial
	interfaceFunc = net.InterfaceAddrs
	fatalFunc     = func(v ...any) { log.Fatal(v...) }
)

// GetOutboundIP dynamically determines the preferred local IP address.
// It first attempts to resolve the standard default route.
// If the network is entirely offline (no internet or default gateway), it falls back
// to sequentially scanning the machine's active network interfaces to find a valid IPv4.
func GetOutboundIP() net.IP {
	// Attempt 1: Dial external IP to allow OS to auto-select the interface with default internet route.
	conn, err := dialFunc("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP
	}

	// Attempt 2: Offline Fallback. Iterate interfaces and pick the first valid non-loopback IPv4 address.
	addrs, err := interfaceFunc()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP
				}
			}
		}
	}

	fatalFunc("Fatal: Failed fetching outbound IP. Ensure you are connected to a network.")
	return nil
}
