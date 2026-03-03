// Package audiostreamer implements streaming capabilities of audio data
package audiostreamer

import (
	"context"
	"encoding/binary"
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
	Queue          deque.Deque[string] // List of tracks
	Interval       time.Duration
	SourceProvider SourceProvider
	Clock          clock.MediaClock
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
		Clock:          *clock.NewMediaClock(),
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
	defer ticker.Stop()

	buffer := make([]byte, 3528)

	for {
		streamer.Clock.UpdateMediaTime()
		if streamer.Clock.GetSentTimeInt64()-streamer.Clock.GetMediaTimeInt64() > streamer.LookAhead {
			<-ticker.C
			continue
		}

		n, err := audioSource.Read(buffer)
		if n == 0 || err != nil {
			logging.Log(logPrefix, "Exiting play loop: n="+strconv.Itoa(n)+" err="+err.Error())
			break
		}

		absoluteStartTime := streamer.Clock.GetStartTimeInt64()
		playAt := streamer.Clock.GetSentTimeInt64()
		envelope := make([]byte, headerSize+n)
		binary.BigEndian.PutUint64(envelope[0:8], uint64(absoluteStartTime))
		binary.BigEndian.PutUint64(envelope[8:16], uint64(playAt))
		copy(envelope[headerSize:], buffer[:n])

		for _, f := range streamer.Followers.Followers {
			streamAudioToFollower(envelope, f)
		}
		durationSent := int64(n) * 1e9 / (audioSource.SampleRate * audioSource.Channels * 2)
		streamer.Clock.AddToSentTime(durationSent)
	}
}

func streamAudioToFollower(buffer []byte, follower follower.Follower) {
	if follower.Conn == nil {
		follower.InitConnection()
	}
	conn := follower.GetConnection()
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
			streamer.Clock = *clock.NewMediaClock()
			streamer.Clock.UpdateStartTime(int(streamer.LookAhead))
			streamer.mu.Unlock()
			streamer.StreamAudioToAll()
		}
	}
}
