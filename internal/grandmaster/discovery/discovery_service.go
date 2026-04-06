// Package discovery implements mDNS for the grandmaster to discover the followers
package discovery

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"ensync/internal/grandmaster/follower"

	"github.com/hashicorp/mdns"
)

const mdnsName = "_ensync._tcp"

type DiscoveryService struct {
	Registry *follower.FollowersRegistry
	NTPPort  string
}

func NewDiscoveryService(registry *follower.FollowersRegistry, ntpPort string) *DiscoveryService {
	return &DiscoveryService{
		Registry: registry,
		NTPPort:  ntpPort,
	}
}

func (ds *DiscoveryService) StartDiscovery() {
	log.SetOutput(io.Discard)

	entriesCh := make(chan *mdns.ServiceEntry, 16)
	go ds.DiscoverFollower(entriesCh)
	go ds.ScanForServers(entriesCh)
}

func (ds *DiscoveryService) ScanForServers(entriesCh chan *mdns.ServiceEntry) {
	params := mdns.DefaultParams(mdnsName)
	params.Entries = entriesCh
	params.DisableIPv6 = false
	params.Timeout = 5 * time.Second
	for {
		err := mdns.Query(params)
		if err != nil {
			fmt.Println("[Discovery] Query error:", err)
		}
		time.Sleep(10 * time.Second)
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
		if _, exists := ds.Registry.Registry[ipAddress]; !exists {
			fmt.Println("[Discovery] Found entry ", ipAddress, endpoint, entry.Port, entry.Name)
			follower.SubscribeFollower(ds.Registry, url, ds.NTPPort)
		}
	}
}
