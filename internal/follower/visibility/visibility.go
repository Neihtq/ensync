// Package visibility implements an mDNS server to be discoverd by the Grandmaster
package visibility

import (
	"os"

	"github.com/hashicorp/mdns"
)

func ExposeMDNS(port int, info []string) *mdns.Server {
	host, _ := os.Hostname()
	service, _ := mdns.NewMDNSService(host, "_ensync._tcp", "", "", port, nil, info)

	server, _ := mdns.NewServer(&mdns.Config{Zone: service})

	return server
}
