// Package discovery implements mDNS for the grandmaster to discover the followers
package discovery

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"ensync/internal/grandmaster/follower"

	"github.com/hashicorp/mdns"
)

const mdnsName = "_ensync._tcp"

type DiscoveryService struct {
	Followers *follower.Followers
	NTPPort   string
}

func NewDiscoveryService(followers *follower.Followers, ntpPort string) *DiscoveryService {
	return &DiscoveryService{
		Followers: followers,
		NTPPort:   ntpPort,
	}
}

func (ds *DiscoveryService) Discover() {
	entriesCh := make(chan *mdns.ServiceEntry, 16)
	go ds.DiscoverFollower(entriesCh)
	go ds.ScanForServers(entriesCh)
}

func (ds *DiscoveryService) ScanForServers(entriesCh chan *mdns.ServiceEntry) {
	for {
		ifaces, err := net.Interfaces()
		if err != nil {
			fmt.Println("Error getting interfaces:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println("Query for Service:", mdnsName)

		for _, ifc := range ifaces {
			if (ifc.Flags&net.FlagLoopback) != 0 || (ifc.Flags&net.FlagUp) == 0 || (ifc.Flags&net.FlagMulticast) == 0 {
				continue
			}

			i := ifc
			params := &mdns.QueryParam{
				Service:             mdnsName,
				Domain:              "",
				Timeout:             2 * time.Second,
				Interface:           &i,
				Entries:             entriesCh,
				WantUnicastResponse: false,
				DisableIPv6:         true,
			}

			go func(p *mdns.QueryParam) {
				_ = mdns.Query(p)
			}(params)
		}

		time.Sleep(2 * time.Second)
	}
}

func (ds *DiscoveryService) DiscoverFollower(entriesCh chan *mdns.ServiceEntry) {
	for entry := range entriesCh {
		if entry == nil || entry.AddrV4 == nil || len(entry.InfoFields) == 0 {
			continue
		}

		if !strings.Contains(entry.Name, "_ensync") {
			continue
		}

		endpoint := entry.InfoFields[0]
		ipAddress := entry.AddrV4.String()
		url := ipAddress + ":" + strconv.Itoa(entry.Port) + endpoint
		fmt.Println("=========≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠")
		fmt.Println("[Discovery] Found entry ", ipAddress, endpoint, entry.Port, entry.Name)
		fmt.Println("=========≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠")
		if _, exists := ds.Followers.Followers[ipAddress]; !exists {
			follower.SubscribeFollower(ds.Followers, url, ds.NTPPort)
		}
	}
}
