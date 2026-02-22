// Package audiostreamer implements streaming capabilities of audio data
package audiostreamer

import (
	"context"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"ensync/internal/grandmaster/clock"
	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/subscription"

	"github.com/gammazero/deque"
)

const (
	logPrefix  = "[AudioStreamer]"
	headerSize = 8
)

type AudioStreamer struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	Subs           *subscription.Subscribers
	Queue          deque.Deque[string] // List of tracks
	Interval       time.Duration
	SourceProvider SourceProvider
	Clock          clock.MediaClock
}

func NewAudioStreamer(
	subs *subscription.Subscribers,
	interval time.Duration,
	sourceProvider SourceProvider,
) *AudioStreamer {
	ctx, cancel := context.WithCancel(context.Background())
	return &AudioStreamer{
		ctx:            ctx,
		cancel:         cancel,
		Subs:           subs,
		Interval:       10 * time.Millisecond,
		SourceProvider: sourceProvider,
	}
}

func (streamer *AudioStreamer) AddToQueue(filePath string) {
	streamer.mu.Lock()
	defer streamer.mu.Unlock()
	streamer.Queue.PushBack(filePath)
}

func (streamer *AudioStreamer) StreamAudioToAll() {
	streamer.mu.Lock()
	filePath := streamer.Queue.PopFront()
	streamer.mu.Unlock()
	logging.Log(logPrefix, "Stream "+filePath)

	audioSource := streamer.SourceProvider.GetSource(filePath)
	ticker := time.NewTicker(streamer.Interval)

	buffer := make([]byte, 3528)

	for range ticker.C {
		n, err := audioSource.Read(buffer)
		if n == 0 || err != nil {
			logging.Log(logPrefix, "Exiting play loop: n="+strconv.Itoa(n)+" err="+err.Error())
			break
		}

		envelope := make([]byte, headerSize+n)
		playAt := streamer.Clock.StampTime(streamer.Interval)
		binary.BigEndian.PutUint64(envelope[:headerSize], uint64(playAt))
		copy(envelope[8:], buffer[:n])

		for _, url := range streamer.Subs.AudioURLs {
			streamAudioToFollower(envelope, url)
		}
	}
}

func streamAudioToFollower(buffer []byte, url string) {
	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conn.Write(buffer)
}

func (streamer *AudioStreamer) StreamAudioToAllLoop(sleepDuration time.Duration, stop chan struct{}) {
	for {
		select {
		case <-stop:
			logging.Log(logPrefix, "Stopping Audio Streamer...")
			return
		default:
			streamer.mu.Lock()
			if streamer.Queue.Len() == 0 {
				streamer.mu.Unlock()
				time.Sleep(sleepDuration)
				continue
			}
			streamer.mu.Unlock()
			streamer.StreamAudioToAll()
		}
	}
}
