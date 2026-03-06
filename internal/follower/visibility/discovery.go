// Package visibility implements an mDNS server to be discoverd by the Grandmaster
package visibility

import (
	"os"

	"github.com/hashicorp/mdns"
)

func ExposeMDNS(port int) {
	host, _ := os.Hostname()
	info := []string{"THE Follower service"}
	service, _ := mdns.NewMDNSService(host, "_ensync._tcp", "", "", port, nil, info)

	server, _ := mdns.NewServer(&mdns.Config{Zone: service})
	defer server.Shutdown()

	select {}
}
