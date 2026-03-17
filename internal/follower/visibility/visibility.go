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

	"github.com/hashicorp/mdns"
)

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

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

func JoinLobby(addr string, port string, endpoint string) error {
	fmt.Println("Joining Lobby...")
	ipAddr := GetOutboundIP().String()
	data := map[string]string{"address": ipAddr, "port": port, "endpoint": endpoint}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post("http://"+addr, "application/json", bytes.NewBuffer(jsonData))
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
