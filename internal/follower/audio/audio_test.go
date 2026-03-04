package audio

import (
	"encoding/binary"
	"fmt"
	"net"
	"slices"
	"testing"
	"time"

	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

func TestReadAboveThreshol(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	audioStream := NewAudioStream(mirrorClock)
	playAt := int64(5)
	mockData := make([]byte, 38400)
	audioStream.WriteToBuffer(mockData, playAt)
	mockBuffer := make([]byte, len(mockData))

	// act
	numBytes, err := audioStream.Read(mockBuffer)
	// assert
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if numBytes != len(mockData) {
		t.Errorf("Read from Buffer above threshold failed: Expected %d number of bytes but received %d number of bytes", len(mockData), numBytes)
	}
}

func TestReadBelowThreshold(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	audioStream := NewAudioStream(mirrorClock)
	playAt := int64(5)
	mockData := []byte{1, 2, 3}
	mockBuffer := make([]byte, len(mockData))
	audioStream.WriteToBuffer(mockData, playAt)

	// act
	numBytes, err := audioStream.Read(mockBuffer)
	// assert
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if numBytes != len(mockData) {
		t.Errorf("Read from Buffer below threshold failed: Expected %d number of bytes but received %d number of bytes", len(mockData), numBytes)
	}
}

func TestWriteToBuffer(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	audioStream := NewAudioStream(mirrorClock)
	playAt := int64(5)
	mockData := []byte{1, 2, 3}
	audioStream.WriteToBuffer(mockData, playAt)

	// act
	chunks := audioStream.chunks

	// assert
	if chunks.Len() != 1 {
		t.Errorf("Writing to Buffer failed: Expected number of chunks %d but is %d", 1, chunks.Len())
	}
	chunk := chunks.Front()
	if !slices.Equal(chunk.data, mockData) || chunk.playAt != playAt {
		t.Errorf("Writing to Buffer failed: Expected data %v and playAt %v but got data %v and playAt %v", mockData, playAt, chunk.data, chunk.playAt)
	}
}

func TestReadEmptyBuffer(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	audioStream := NewAudioStream(mirrorClock)
	mockBuffer := []byte{}

	// act
	numBytes, err := audioStream.Read(mockBuffer)
	// assert
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if numBytes > 0 {
		t.Errorf("Read from empty Buffer failed: Expected %d number of bytes but received %d number of bytres", 0, numBytes)
	}
}

func TestAudioStreamHasEmptyBufferAgain(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	audioStream := NewAudioStream(mirrorClock)
	playAt := int64(5)
	mockData := make([]byte, 38400)
	audioStream.WriteToBuffer(mockData, playAt)
	mockBuffer := make([]byte, len(mockData))

	// empty buffer and flip isBuffering flag again
	fmt.Println(len(mockBuffer))
	audioStream.Read(mockBuffer)
	fmt.Println(len(mockBuffer))

	// act
	_, err := audioStream.Read(mockBuffer)
	// assert
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if !audioStream.isBuffering {
		t.Errorf("Test Audio stream has empty buffer again failed: Expected isBuffering to be true but was %v", audioStream.isBuffering)
	}
}

func TestAudioStreamDoesNotPlayWhenItIsNotTime(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	audioStream := NewAudioStream(mirrorClock)
	playAt := int64(time.Now().Add(1 * time.Hour).UnixNano())
	mockData := make([]byte, 38400)
	audioStream.WriteToBuffer(mockData, playAt)
	mockBuffer := make([]byte, len(mockData))

	// empty buffer and flip isBuffering flag again
	audioStream.Read(mockBuffer)

	// act
	numBytes, err := audioStream.Read(mockBuffer)
	// assert
	if err != nil {
		t.Errorf("Read failed. %s", err.Error())
	}
	if numBytes > 38400 {
		t.Errorf("PlayAt test failed. Expected 38400 bytes but got %d", numBytes)
	}
}

func sendTestUDPPacket(t *testing.T, url string, stop chan struct{}) {
	fmt.Println("URL DIAL " + url)
	conn, err := net.Dial("udp", url)
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}

	header := make([]byte, 8)
	binary.BigEndian.PutUint64(header, uint64(time.Now().Nanosecond()))
	payload := []byte("test message")
	envelope := append(header, payload...)

	_, err = conn.Write(envelope)
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}
	conn.Close()

	time.Sleep(200 * time.Millisecond)
	close(stop)
}

func TestLaunchAudioServer(t *testing.T) {
	port := ":9011"
	ipProvider := middleware.MockIPProvider{FakeIP: []byte{127, 0, 0, 1}}
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})

	go LaunchAudioServer(port, ipProvider, mirrorClock, stop)
	time.Sleep(100 * time.Millisecond)

	address := ipProvider.GetIP().String() + port

	sendTestUDPPacket(t, address, stop)
}

func TestExpose(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	audioStream := NewAudioStream(mirrorClock)
	port := "9001"
	ipProvider := middleware.MockIPProvider{FakeIP: []byte{127, 0, 0, 1}}
	address := ipProvider.GetIP().String() + ":" + port
	stop := make(chan struct{})

	// act
	go expose(address, audioStream, mirrorClock, stop)
	time.Sleep(100 * time.Millisecond)

	sendTestUDPPacket(t, address, stop)

	// assert
	audioStream.mu.Lock()
	defer audioStream.mu.Unlock()
	if audioStream.chunks.Len() < 1 {
		t.Errorf("Test expose() failed: Message not written into stream.")
	}
}

func TestCheckPlayerErrorHasNoErrors(t *testing.T) {
	// arrange
	mockErrorReporter := MockErrorReporter{}
	sleepInterval := 1 * time.Nanosecond

	// act
	go checkPlayerError(&mockErrorReporter, sleepInterval)
}

func TestCheckPlayerErrorHasErrors(t *testing.T) {
	// arrange
	mockErrorReporter := MockErrorReporter{mockErr: fmt.Errorf("Some error")}
	sleepInterval := 1 * time.Second

	// act
	go checkPlayerError(&mockErrorReporter, sleepInterval)
}
