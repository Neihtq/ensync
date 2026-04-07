package discovery

import (
	"net"
	"testing"
	"time"

	"ensync/internal/follower/controlplane"
	"ensync/internal/follower/mirrorclock"
	"ensync/internal/grandmaster/follower"

	"github.com/hashicorp/mdns"
)

func TestDiscoverFollower(t *testing.T) {
	t.Skip("Skipping mDNS test")

	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})
	audioPort := ":4222"
	cp := controlplane.NewControlPlaneService(mirrorClock, audioPort, stop)
	cpPort := ":7777"
	go cp.StartService(cpPort)
	time.Sleep(20 * time.Millisecond)

	heartbeatPort := ":65533"
	registry := follower.NewFollowersRegistry(heartbeatPort)
	ntpPort := ":9111"

	addrV4 := "127.0.0.1"
	port := 7777
	infoFields := []string{"/connections"}
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	entry := mdns.ServiceEntry{
		AddrV4:     net.ParseIP(addrV4),
		Port:       port,
		InfoFields: infoFields,
	}
	entriesCh <- &entry
	close(entriesCh)

	// act
	discoveryService := NewDiscoveryService(registry, ntpPort)
	discoveryService.DiscoverFollower(entriesCh)
}

func TestDiscover(t *testing.T) {
	t.Skip("Skipping mDNS test")
	// arrange
	ntpPort := ":9999"
	registry := follower.NewFollowersRegistry(ntpPort)

	// act
	discoveryService := NewDiscoveryService(registry, ntpPort)
	discoveryService.StartDiscovery()
}
