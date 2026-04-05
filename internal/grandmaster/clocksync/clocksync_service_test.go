package clocksync

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func TestExposeNTP(t *testing.T) {
	// arrange
	stop := make(chan struct{})
	serverAddrStr := "127.0.0.1:10123"

	// ct
	clockSyncService := NewClockSyncService(serverAddrStr)
	go func() {
		if err := clockSyncService.ExposeNTP(stop); err != nil {
			t.Errorf("ExposeNTP failed: Server exited with error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	clientConn, err := net.Dial("udp", serverAddrStr)
	if err != nil {
		t.Fatalf("ExposeNTP failed: %v", err)
	}
	defer clientConn.Close()

	t1 := uint64(time.Now().UnixNano())
	sendBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(sendBuffer, t1)

	_, err = clientConn.Write(sendBuffer)
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}

	recvBuffer := make([]byte, 24)
	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := clientConn.Read(recvBuffer)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if n != 24 {
		t.Errorf("Expected 24 bytes, got %d", n)
	}

	// assert

	returnedT1 := binary.BigEndian.Uint64(recvBuffer[:8])
	t2 := binary.BigEndian.Uint64(recvBuffer[8:16])
	t3 := binary.BigEndian.Uint64(recvBuffer[16:24])

	if returnedT1 != t1 {
		t.Errorf("Original timestamp mismatch. Sent %d, got %d", t1, returnedT1)
	}

	if t2 == 0 || t3 == 0 {
		t.Error("Server timestamps (T2/T3) should not be zero")
	}

	if t3 < t2 {
		t.Error("Server Transmit time (T3) cannot be before Receive time (T2)")
	}

	close(stop)
}
