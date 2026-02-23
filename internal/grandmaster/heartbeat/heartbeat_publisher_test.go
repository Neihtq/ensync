package heartbeat

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"ensync/internal/grandmaster/follower"
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

	follow := follower.Follower{HeartbeatURL: serverAddr}
	followers := follower.NewFollowers()
	followers.Followers[serverAddr] = follow
	timeNow := uint64(time.Now().UnixNano())

	// act
	heartbeatPublisher := &HeartbeatPublisher{Followers: followers}
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

func TestHeartBeatLoop(t *testing.T) {
	// arrange
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}

	serverAddr := conn.LocalAddr().String()
	follows := map[string]follower.Follower{
		serverAddr: {HeartbeatURL: serverAddr},
	}
	followers := follower.Followers{Followers: follows}
	heartbeatPublisher := &HeartbeatPublisher{Followers: &followers}

	interval := 1 * time.Nanosecond
	stop := make(chan struct{})

	// act
	go heartbeatPublisher.HeartbeatLoop(interval, stop)

	time.Sleep(100 * time.Millisecond)
	close(stop)
}
