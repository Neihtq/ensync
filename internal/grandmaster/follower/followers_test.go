package follower

import (
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

	registry := NewFollowersRegistry()
	endpoint := "/connections"
	url := "127.0.0.1"
	ntpPort := ":9111"

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
