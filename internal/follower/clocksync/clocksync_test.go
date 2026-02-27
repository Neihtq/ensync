package clocksync

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"ensync/internal/follower/mirrorclock"
)

func TestSendNTPRequest(t *testing.T) {
	// arrange
	clock := mirrorclock.NewMirrorClock()
	serverURL := "127.0.0.1:0"
	addr, _ := net.ResolveUDPAddr("udp", serverURL)
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	timeNow := uint64(time.Now().UnixNano())
	clockSync := NewClockSync(clock, conn.LocalAddr().String())

	received := make(chan []byte)
	go func() {
		buffer := make([]byte, 8)
		_, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic("Error receigin UDP message: " + err.Error())
		}
		received <- buffer
	}()

	// act
	err := clockSync.SendNTPRequest()
	// arrange
	if err != nil {
		t.Fatalf("SendNTPRequest failed: Threw error %s", err.Error())
	}
	select {
	case msg := <-received:
		timestamp := binary.BigEndian.Uint64(msg)
		if timestamp <= timeNow {
			t.Fatalf("SendNTPRequest failed: Sent timestamp should be higher than %d but is %d", timeNow, timestamp)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("SendNTPRequest timed out: No UDP packed received.")
	}
}

func TestReceiveNTPPackets(t *testing.T) {
	// arrange
	clock := mirrorclock.NewMirrorClock()
	serverURL := "127.0.0.1:9990"
	clockSync := NewClockSync(clock, serverURL)

	stop := make(chan struct{})

	// act
	go clockSync.ReceiveNTPPackets(stop)
	time.Sleep(100 * time.Millisecond)

	// assert
	clockSync.mu.Lock()
	targetAddr := clockSync.ListeningAddr
	clockSync.mu.Unlock()
	if targetAddr == "" {
		t.Fatalf("Target Address not set!")
	}
	addr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		t.Fatalf("ReceiveNTPPackets failed: Could not resolve address for url %s", serverURL)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatalf("ReceiveNTPPackets failed: Could not establish UDP connection to %s", conn.LocalAddr().String())
	}
	packet := make([]byte, 24)
	testTime := uint64(time.Now().UnixNano())
	binary.BigEndian.PutUint64(packet[:8], testTime)
	binary.BigEndian.PutUint64(packet[8:16], testTime)
	binary.BigEndian.PutUint64(packet[16:], testTime)
	_, err = conn.Write(packet)
	conn.Close()
	time.Sleep(50 * time.Millisecond)
	close(stop)

	if err != nil {
		t.Fatalf("ReceiveNTPPackets failed: Failed to send test message: %s", err.Error())
	}
	ntpOffset, delay := clockSync.Clock.GetNTPStats()
	if ntpOffset < 10 {
		t.Fatalf("ReceiveNTPPackets failed: Clock offset should be synced to 0 but is %d", ntpOffset)
	}
	if delay < 10 {
		t.Fatalf("ReceiveNTPPackets failed: Clock delay should be synced to 0 but is %d", delay)
	}
}
