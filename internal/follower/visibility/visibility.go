// Package visibility implements an mDNS server to be discoverd by the Grandmaster
package visibility

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/mdns"
)

func ExposeMDNS(port int, info []string) (*mdns.Server, error) {
	host, _ := os.Hostname()
	cleanHost := strings.TrimSuffix(host, ".local")
	service, err := mdns.NewMDNSService(cleanHost, "_ensync._tcp", "local.", fmt.Sprintf("%s.local.", cleanHost), port, nil, info)
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
