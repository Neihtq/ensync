package discovery

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ensync/internal/follower/controlplane"
	"ensync/internal/follower/mirrorclock"
	"ensync/internal/grandmaster/follower"

	"github.com/hashicorp/mdns"
)

func TestDiscoverFollower(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})
	audioPort := ":4222"
	cp := controlplane.NewControlPlaneService(mirrorClock, audioPort, stop)
	cpPort := ":7777"
	go cp.StartService(cpPort)
	time.Sleep(20 * time.Millisecond)

	followers := follower.NewFollowers()
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
	discoveryService := NewDiscoveryService(followers, ntpPort)
	discoveryService.DiscoverFollower(entriesCh)
}

func TestDiscover(t *testing.T) {
	// arrange
	followers := follower.NewFollowers()
	ntpPort := ":9999"

	// act
	discoveryService := NewDiscoveryService(followers, ntpPort)
	discoveryService.Discover()
}

func TestLobby(t *testing.T) {
	// arrange
	followers := follower.NewFollowers()
	stop := make(chan struct{})
	port := ":11111"

	// assert
	dl := NewDiscoveryLobby(followers, stop)
	dl.OpenLobby(port)

	close(stop)
}

func TestJoinLobby(t *testing.T) {
	// arrange
	stop := make(chan struct{})
	followers := follower.NewFollowers()
	visitorPort := ":11112"
	ipAddress := "127.0.0.1"
	jsonBody := []byte(`{"address":"` + ipAddress + `","port":"` + visitorPort + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/visitor", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	// act
	dl := NewDiscoveryLobby(followers, stop)
	dl.JoinLobby(writer, request)

	// assert
	if dl.visitors[ipAddress] != visitorPort {
		t.Errorf("Visitor creation failed: Expected port %s for ip address %s but received %s", visitorPort, ipAddress, dl.visitors[ipAddress])
	}
	if writer.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", writer.Code)
	}
}

func TestTransferVisitorsToFollowers(t *testing.T) {
	stop := make(chan struct{})
	followers := follower.NewFollowers()
	visitorPort := ":11113"
	ipAddress := "127.0.0.1"
	jsonBody := []byte(`{"address":"` + ipAddress + `","port":"` + visitorPort + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/visitor", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	dl := NewDiscoveryLobby(followers, stop)
	dl.JoinLobby(writer, request)

	// act
	dl.TransferVisitorsToFollowers()
}
