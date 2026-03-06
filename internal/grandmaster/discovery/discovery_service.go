// Package discovery implements mDNS for the grandmaster to discover the followers
package discovery

import (
	"fmt"
	"strconv"
	"time"

	"ensync/internal/grandmaster/follower"

	"github.com/hashicorp/mdns"
)

const mdnsName = "_ensync._tcp."

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
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go ds.DiscoverFollower(entriesCh)
	go ds.ScanForServers(entriesCh)
}

func (ds *DiscoveryService) ScanForServers(entriesCh chan *mdns.ServiceEntry) {
	for {
		params := &mdns.QueryParam{
			Service:             mdnsName,
			Domain:              "local",
			Timeout:             2 * time.Second,
			Entries:             entriesCh,
			WantUnicastResponse: false,
			DisableIPv6:         true,
		}

		fmt.Println("Query for Service:", params.Service)
		err := mdns.Query(params)
		if err != nil {
			fmt.Println("mDNS query error:", err)
		}
		time.Sleep(2 * time.Second)
	}
}

func (ds *DiscoveryService) DiscoverFollower(entriesCh chan *mdns.ServiceEntry) {
	for entry := range entriesCh {
		endpoint := entry.InfoFields[0]
		ipAddress := entry.AddrV4.String()
		url := ipAddress + ":" + strconv.Itoa(entry.Port) + endpoint
		fmt.Println("=========≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠")
		fmt.Println("[Discovery] Found entry ", entry.AddrV4, endpoint, entry.Port, entry.Name)
		fmt.Println("=========≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠≠")
		if _, exists := ds.Followers.Followers[ipAddress]; !exists {
			follower.SubscribeFollower(ds.Followers, url, ds.NTPPort)
		}
	}
}
