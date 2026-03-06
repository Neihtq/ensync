// Package visibility implements an mDNS server to be discoverd by the Grandmaster
package visibility

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/mdns"
)

func getLocalIPs() []net.IP {
	var ips []net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, i := range ifaces {
		if (i.Flags&net.FlagLoopback) == 0 && (i.Flags&net.FlagUp) != 0 {
			addrs, _ := i.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						ips = append(ips, ipnet.IP)
					}
				}
			}
		}
	}
	// Fallback to localhost if no external IP found
	if len(ips) == 0 {
		ips = append(ips, net.ParseIP("127.0.0.1"))
	}
	return ips
}

func ExposeMDNS(port int, info []string) (*mdns.Server, error) {
	host, _ := os.Hostname()
	cleanHost := strings.TrimSuffix(host, ".local")
	ips := getLocalIPs()
	service, err := mdns.NewMDNSService(cleanHost, "_ensync._tcp", "local.", fmt.Sprintf("%s.local.", cleanHost), port, ips, info)
	if err != nil {
		fmt.Println("Failed mDNS Service initialization ", err.Error())
		return nil, err
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		fmt.Println("Failed mDNS Server initialization ", err.Error())
		return nil, err
	}

	return server, nil
}
