package middleware

import (
	"net"
	"testing"
	"time"
)

func startTestServer(t *testing.T, port string, stop chan struct{}) {
	err := expose(port, stop)
	if err != nil {
		t.Errorf("Server failed : %v", err)
	}
}

func sendTestUDPPacket(t *testing.T, port string, stop chan struct{}) {
	conn, err := net.Dial("udp", "127.0.0.1"+port)
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
	stop := make(chan struct{})
	port := ":9000"

	go startTestServer(t, port, stop)
	time.Sleep(100 * time.Millisecond)

	sendTestUDPPacket(t, port, stop)
}

func TestSubscribeAndExpose(t *testing.T) {
	port := ":9000"
	stop := make(chan struct{})
	ipProvider := MockIPProvider{FakeIP: []byte{127, 0, 0, 1}}
	server := StartMockHTTPServer(t)
	endpointProvider := MockEndpointProvider{server.URL}

	go SubscribeAndExpose(port, stop, ipProvider, endpointProvider)
	time.Sleep(100 * time.Millisecond)

	sendTestUDPPacket(t, port, stop)
}
