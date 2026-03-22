// Package audiostreamer implements streaming capabilities of audio data
package audiostreamer

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"
	"time"

	"ensync/internal/grandmaster/clock"
	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/logging"

	"github.com/gammazero/deque"
)

const (
	logPrefix  = "[AudioStreamer]"
	headerSize = 16
)

type AudioStreamer struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	Followers      *follower.Followers
	TrackQueue     deque.Deque[string] // List of tracks
	Interval       time.Duration
	SourceProvider SourceProvider
	MediaClock     clock.MediaClock
	LookAhead      int64
}

func NewAudioStreamer(
	followers *follower.Followers,
	interval time.Duration,
	lookAhead int64,
	sourceProvider SourceProvider,
) *AudioStreamer {
	ctx, cancel := context.WithCancel(context.Background())
	return &AudioStreamer{
		ctx:            ctx,
		cancel:         cancel,
		Followers:      followers,
		Interval:       interval,
		LookAhead:      lookAhead,
		SourceProvider: sourceProvider,
		MediaClock:     *clock.NewMediaClock(),
	}
}

func (streamer *AudioStreamer) AddToQueue(filePath string) {
	streamer.mu.Lock()
	defer streamer.mu.Unlock()
	streamer.TrackQueue.PushBack(filePath)
}

func (streamer *AudioStreamer) StreamAudioToAll() {
	streamer.mu.Lock()
	filePath := streamer.TrackQueue.PopFront()
	streamer.mu.Unlock()
	logging.Log(logPrefix, "Stream "+filePath)

	audioSource := streamer.SourceProvider.GetSource(filePath)
	ticker := time.NewTicker(streamer.Interval)
	defer ticker.Stop()

	buffer := make([]byte, 3528)

	for {
		streamer.MediaClock.UpdateMediaTime()
		if streamer.MediaClock.GetSentTimeInt64()-streamer.MediaClock.GetMediaTimeInt64() > streamer.LookAhead {
			<-ticker.C
			continue
		}

		dataSize, err := audioSource.Read(buffer)
		if dataSize == 0 || err != nil {
			logging.Log(logPrefix, "Exiting play loop: n="+strconv.Itoa(dataSize)+" err="+err.Error())
			break
		}

		envelope := streamer.prepareEnvelope(buffer, dataSize)

		for _, f := range streamer.Followers.Followers {
			streamAudioToFollower(envelope, f)
		}

		durationSent := int64(dataSize) * 1e9 / (audioSource.SampleRate * audioSource.Channels * 2)
		streamer.MediaClock.AddToSentTime(durationSent)
	}
}

func (streamer *AudioStreamer) prepareEnvelope(buffer []byte, packetSize int) []byte {
	absoluteStartTime := streamer.MediaClock.GetStartTimeInt64()
	playAt := streamer.MediaClock.GetSentTimeInt64()
	envelope := make([]byte, headerSize+packetSize)
	binary.BigEndian.PutUint64(envelope[0:8], uint64(absoluteStartTime))
	binary.BigEndian.PutUint64(envelope[8:16], uint64(playAt))
	copy(envelope[headerSize:], buffer[:packetSize])

	return envelope
}

func streamAudioToFollower(buffer []byte, follower *follower.Follower) {
	conn := follower.GetConnection()
	if conn == nil {
		fmt.Println("Initiate streaming connection")
		follower.InitConnection()
		conn = follower.GetConnection()
	}
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
			if streamer.TrackQueue.Len() == 0 {
				streamer.mu.Unlock()
				time.Sleep(sleepDuration)
				continue
			}
			streamer.MediaClock = *clock.NewMediaClock()
			streamer.MediaClock.UpdateStartTime(int(streamer.LookAhead))
			streamer.mu.Unlock()
			streamer.StreamAudioToAll()
		}
	}
}
