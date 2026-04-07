package follower

import (
	"net"
	"testing"
	"time"

	"ensync/internal/common/netutil"
	"ensync/internal/follower/controlplane"
	"ensync/internal/follower/mirrorclock"
)

func TestSubscribeFollower(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})

	audioPort := ":4222"
	cp := controlplane.NewControlPlaneService(mirrorClock, audioPort, stop)
	cpPort := ":7777"

	go cp.StartService(cpPort)
	time.Sleep(20 * time.Millisecond)

	endpoint := "/connections"
	url := "127.0.0.1"
	ntpPort := ":9111"
	registry := NewFollowersRegistry(ntpPort)

	tcpAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0"+ntpPort)
	tcpConn, _ := net.ListenTCP("tcp", tcpAddr)
	defer tcpConn.Close()

	// act
	err := SubscribeFollower(registry, url+cpPort+endpoint, ntpPort)
	close(stop)

	// assert
	if err != nil {
		t.Errorf("Error subscribing follower: %s", err.Error())
	}

	if len(registry.Registry) == 0 {
		t.Errorf("0 registered Followers.")
	}

	ipAddr := netutil.GetOutboundIP().String()
	expected := ipAddr + audioPort
	registered, exists := registry.Registry[ipAddr]
	if !exists {
		t.Errorf("Register Follower failed: Follower not found")
	}
	if registered.AudioURL != expected {
		t.Errorf("Expected registered follower AudioURL %s but was %s", expected, registered.AudioURL)
	}
}

func TestNewFollower(t *testing.T) {
	f := NewFollower("192.168.1.5:8080")
	if f.AudioURL != "192.168.1.5:8080" {
		t.Errorf("expected AudioURL '192.168.1.5:8080', got '%s'", f.AudioURL)
	}
	if f.Conn != nil {
		t.Errorf("expected Conn to be nil initially")
	}
}

func TestFollower_GetConnection_Nil(t *testing.T) {
	f := NewFollower("192.168.1.5:8080")
	conn := f.GetConnection()
	if conn != nil {
		t.Errorf("expected nil connection before InitConnection")
	}
}

func TestNewFollowersRegistry(t *testing.T) {
	heartbeatPort := ":65533"
	registry := NewFollowersRegistry(heartbeatPort)
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(registry.Registry) != 0 {
		t.Errorf("expected empty registry, got %d entries", len(registry.Registry))
	}
}

func TestFollowersRegistry_GetAllFollowers(t *testing.T) {
	heartbeatPort := ":65533"
	registry := NewFollowersRegistry(heartbeatPort)

	followers := registry.GetAllFollowers()
	if len(followers) != 0 {
		t.Errorf("expected 0 followers, got %d", len(followers))
	}

	registry.RegisterFollower("10.0.0.1", "8080")
	registry.RegisterFollower("10.0.0.2", "8080")

	followers = registry.GetAllFollowers()
	if len(followers) != 2 {
		t.Errorf("expected 2 followers, got %d", len(followers))
	}
}

func TestFollower_InitConnection(t *testing.T) {
	// Use a local address for testing
	f := NewFollower("127.0.0.1:9999")

	f.InitConnection()

	conn := f.GetConnection()
	if conn == nil {
		t.Fatal("expected connection to be initialized")
	}
	defer conn.Close()

	// Initializing again should not change anything or panic
	oldConn := conn
	f.InitConnection()
	if f.GetConnection() != oldConn {
		t.Error("expected second InitConnection to be a no-op")
	}
}
