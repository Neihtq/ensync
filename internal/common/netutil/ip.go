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

func GetOutboundIP() net.IP {
	conn, err := dialFunc("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP
	}

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
