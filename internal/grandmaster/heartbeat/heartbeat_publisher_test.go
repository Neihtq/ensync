package heartbeat

import (
	"net"
	"testing"
	"time"

	"ensync/internal/grandmaster/subscription"
)

func TestHeartbeatPublisherProcessesEachUrl(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()
	received := make(chan string)
	go func() {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		received <- string(buf[:n])
	}()

	urls := []string{serverAddr}
	subscribers := subscription.Subscribers{Urls: urls}

	heartbeatPublisher := &HeartbeatPublisher{Subs: &subscribers}
	heartbeatPublisher.SendHeartbeatToAll()

	select {
	case msg := <-received:
		if msg != "heartbeat" {
			t.Errorf("Expected heartbeat, got %s", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Test timed out: No UDP packet received")
	}
}
