package heartbeat

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"ensync/internal/grandmaster/subscription"
)

func TestHeartbeatPublisherProcessesEachUrl(t *testing.T) {
	// arrange
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
	received := make(chan []byte)
	go func() {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		received <- buf[:n]
	}()

	urls := []string{serverAddr}
	subscribers := subscription.Subscribers{HeartbeatURLs: urls}

	timeNow := uint64(time.Now().UnixNano())

	// act
	heartbeatPublisher := &HeartbeatPublisher{Subs: &subscribers}
	heartbeatPublisher.SendHeartbeatToAll()

	// assert
	select {
	case msg := <-received:
		timestamp := binary.BigEndian.Uint64(msg)
		if timestamp <= timeNow {
			t.Errorf("Failed Heartbeat Publishing: Timestamp should be higher than %d but is %d", timeNow, timestamp)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Test timed out: No UDP packet received")
	}
}
