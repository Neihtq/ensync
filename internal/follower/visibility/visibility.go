// Package visibility implements an mDNS server to be discoverd by the Grandmaster
package visibility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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

func JoinLobby(addr string, cpPort string, endpoint string) error {
	address := "http://" + addr
	fmt.Println("Joining Lobby to", address)
	ipAddr := netutil.GetOutboundIP().String()
	data := map[string]string{"address": ipAddr, "port": cpPort, "endpoint": endpoint}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(address, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	fmt.Println("Joining Lobby succeeded")

	return nil
}
