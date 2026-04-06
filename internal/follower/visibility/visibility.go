// Package visibility implements an mDNS server to be discoverd by the Grandmaster
package visibility

import (
	"fmt"
	"net"
	"os"
	"strings"

	"ensync/internal/common/netutil"

	"github.com/hashicorp/mdns"
)

const mDNSServiceName = "_ensync._tcp"

func ExposeMDNS(port int, info []string) (*mdns.Server, error) {
	host, _ := os.Hostname()
	cleanHost := strings.TrimSuffix(host, ".local")
	ipAddr := netutil.GetOutboundIP()
	service, err := mdns.NewMDNSService(
		cleanHost,
		mDNSServiceName,
		"local.",
		fmt.Sprintf("%s.local.", cleanHost),
		port,
		[]net.IP{ipAddr},
		info,
	)
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
