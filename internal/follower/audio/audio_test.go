package audio

import (
	"fmt"
	"net"
	"slices"
	"testing"
	"time"

	"ensync/internal/follower/middleware"
)

func TestAudioStream(t *testing.T) {
	audioStream := NewAudioStream()

	// Test Read empty buffer
	mockBuffer := []byte{}
	numBytes, err := audioStream.Read(mockBuffer)
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if numBytes > 0 {
		t.Errorf("Read from empty Buffer failed: Expected %d number of bytes but received %d number of bytres", 0, numBytes)
	}

	// Test WriteToBuffer
	mockData := []byte{1, 2, 3}
	audioStream.WriteToBuffer(mockData)
	if !slices.Equal(audioStream.data, mockData) {
		t.Errorf("Writing to Buffer failed: Expected %v but received %v", mockData, audioStream.data)
	}

	// Test Read below threshold
	mockBuffer = make([]byte, len(mockData))
	numBytes, err = audioStream.Read(mockBuffer)
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if numBytes > 0 {
		t.Errorf("Read from Buffer below threshold failed: Expected %d number of bytes but received %d number of bytres", len(mockData), numBytes)
	}

	// Test Read above threshold
	mockData = make([]byte, 38400)
	audioStream.WriteToBuffer(mockData)
	mockBuffer = make([]byte, len(mockData))
	numBytes, err = audioStream.Read(mockBuffer)
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if numBytes != len(mockData) {
		t.Errorf("Read from Buffer above threshold failed: Expected %d number of bytes but received %d number of bytres", len(mockData), numBytes)
	}
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

func TestLaunchAudioServer(t *testing.T) {
	port := "9001"
	ipProvider := middleware.MockIPProvider{FakeIP: []byte{127, 0, 0, 1}}
	stop := make(chan struct{})

	go LaunchAudioServer(port, ipProvider, stop)
	time.Sleep(100 * time.Millisecond)

	address := ipProvider.GetIP().String() + ":" + port
	sendTestUDPPacket(t, address, stop)
}
