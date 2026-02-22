package heartbeat

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

func startMockHTTPServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Errorf("Expected POST request, got %s", request.Method)
		}

		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json header, got %s", request.Header.Get("Content-Type"))
		}

		writer.WriteHeader(http.StatusNoContent)
	}))
}

func sendTestUDPPacket(t *testing.T, url string, stop chan struct{}) {
	fmt.Println("URL DIAL " + url)
	conn, err := net.Dial("udp", url)
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}

	message := []byte("test message")
	_, err = conn.Write(message)
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}
	conn.Close()

	time.Sleep(100 * time.Millisecond)
	close(stop)
}

func TestExpose(t *testing.T) {
	// arrange
	stop := make(chan struct{})
	heartbeatPort := "9100"
	ipProvider := middleware.MockIPProvider{FakeIP: []byte{127, 0, 0, 1}}
	mirrorClock := mirrorclock.NewMirrorClock()
	heartbeatReceiver := NewHeartbeatReceiver(heartbeatPort, ipProvider, mirrorClock)

	// act & assert
	go heartbeatReceiver.expose(stop)
	time.Sleep(100 * time.Millisecond)

	sendTestUDPPacket(t, heartbeatReceiver.URL, stop)
}

func TestSubscribeAndExpose(t *testing.T) {
	// arrange
	audioPort := "9000"
	heartbeatPort := "9001"
	stop := make(chan struct{})
	ipProvider := middleware.MockIPProvider{FakeIP: []byte{127, 0, 0, 1}}
	server := startMockHTTPServer(t)
	endpointProvider := middleware.MockEndpointProvider{FakeEndpoint: server.URL}

	mirrorClock := mirrorclock.NewMirrorClock()
	heartbeatReceiver := NewHeartbeatReceiver(heartbeatPort, ipProvider, mirrorClock)

	// act
	go heartbeatReceiver.SubscribeAndExpose(audioPort, stop, endpointProvider)
	time.Sleep(100 * time.Millisecond)

	// assert
	address := ipProvider.GetIP().String() + ":" + heartbeatPort
	sendTestUDPPacket(t, address, stop)
}
