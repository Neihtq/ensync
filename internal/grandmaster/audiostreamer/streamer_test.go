package audiostreamer

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/queue"
	"ensync/internal/grandmaster/sourceprovider"
)

const (
	duration  = 1 * time.Nanosecond
	lookAhead = int64(10 * time.Second)
)

func prepareTestFixtures(t *testing.T) string {
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

	return serverAddr
}

func TestStreamAudioToAll(t *testing.T) {
	// arrange
	serverAddr := prepareTestFixtures(t)
	filePath := "../../../assets/test_file.mp3"
	mockSourceProvider := sourceprovider.MockSourceProvider{}
	follow := follower.Follower{AudioURL: serverAddr}
	followers := follower.NewFollowers()
	followers.Followers[serverAddr] = &follow
	trackQueue := queue.NewTrackQueue()

	// act
	audioStreamer := NewAudioStreamer(followers, duration, lookAhead, &mockSourceProvider, duration, trackQueue)
	audioStreamer.AddToQueue(filePath)
	audioStreamer.StreamAudioToAll()

	// assert
	queueLength := audioStreamer.TrackQueue.Len()
	if queueLength > 0 {
		t.Fatal("Failed StreamAudioToAll: Queue should be empt but had length " + strconv.Itoa(queueLength))
	}
}

func TestStreamAudioToAllLoop(t *testing.T) {
	// arrange
	serverAddr := prepareTestFixtures(t)
	mockSourceProvider := sourceprovider.MockSourceProvider{}
	follow := follower.Follower{AudioURL: serverAddr}
	followers := follower.NewFollowers()
	followers.Followers[serverAddr] = &follow
	trackQueue := queue.NewTrackQueue()

	ctx, cancel := context.WithCancel(context.Background())

	audioStreamer := NewAudioStreamer(followers, duration, lookAhead, &mockSourceProvider, duration, trackQueue)
	audioStreamer.ctx = ctx
	audioStreamer.cancel = cancel

	filePath := "../../../assets/test_file.mp3"
	audioStreamer.AddToQueue(filePath)
	stop := make(chan struct{})

	// act
	go audioStreamer.StreamAudioToAllLoop(stop)

	time.Sleep(100 * time.Millisecond)
	close(stop)
}
