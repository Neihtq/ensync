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
	"ensync/internal/grandmaster/queue"
	"ensync/internal/grandmaster/sourceprovider"
)

const (
	logPrefix  = "[AudioStreamer]"
	headerSize = 20
)

type AudioStreamer struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	Followers *follower.FollowersRegistry

	TrackQueue        *queue.TrackQueue
	StreamingInterval time.Duration
	SourceProvider    sourceprovider.SourceProvider
	MediaClock        clock.MediaClock
	LookAhead         int64
	SleepInterval     time.Duration
	OnTrackChanged    func(trackID string)
}

func NewAudioStreamer(
	followers *follower.FollowersRegistry,
	interval time.Duration,
	lookAhead int64,
	sourceProvider sourceprovider.SourceProvider,
	sleepInterval time.Duration,
	trackQueue *queue.TrackQueue,
) *AudioStreamer {
	ctx, cancel := context.WithCancel(context.Background())
	return &AudioStreamer{
		ctx:               ctx,
		cancel:            cancel,
		Followers:         followers,
		TrackQueue:        trackQueue,
		StreamingInterval: interval,
		LookAhead:         lookAhead,
		SourceProvider:    sourceProvider,
		MediaClock:        *clock.NewMediaClock(),
		SleepInterval:     sleepInterval,
	}
}

func (streamer *AudioStreamer) AddToQueue(filePath string) {
	streamer.TrackQueue.PushBack(filePath)
}

func (streamer *AudioStreamer) SetCallbackHook(callback func(trackID string)) {
	streamer.OnTrackChanged = callback
}

func (streamer *AudioStreamer) StreamAudioToAll(stop chan struct{}) {
	trackID := streamer.TrackQueue.PopFront()
	streamer.TrackQueue.SetNowPlaying(trackID)
	logging.Log(logPrefix, "Stream "+trackID)
	if streamer.OnTrackChanged != nil {
		streamer.OnTrackChanged(trackID)
	}

	audioSource, err := streamer.SourceProvider.GetSource(trackID)
	if err != nil {
		logging.Log(logPrefix, "Error preparing track: "+err.Error())
		return
	}
	defer audioSource.Close()

	ticker := time.NewTicker(streamer.StreamingInterval)
	defer ticker.Stop()

	buffer := make([]byte, 3528)

	for {
		select {
		case <-stop:
			logging.Log(logPrefix, "Cancelling song.")
			return
		default:
			streamer.MediaClock.UpdateMediaTime()
			if streamer.MediaClock.GetSentTimeInt64()-streamer.MediaClock.GetMediaTimeInt64() > streamer.LookAhead {
				<-ticker.C
				continue
			}

			dataSize, err := audioSource.Read(buffer)
			if dataSize == 0 || err != nil {
				logging.Log(logPrefix, "Exiting play loop: n="+strconv.Itoa(dataSize)+" err="+err.Error())
				return
			}

			envelope := streamer.prepareEnvelope(buffer, dataSize, int(audioSource.SampleRate))

			for _, f := range streamer.Followers.Registry {
				streamAudioToFollower(envelope, f)
			}

			durationSent := int64(dataSize) * 1e9 / (audioSource.SampleRate * audioSource.Channels * 2)
			streamer.MediaClock.AddToSentTime(durationSent)

			time.Sleep(2 * time.Millisecond)
		}
	}
}

func (streamer *AudioStreamer) prepareEnvelope(buffer []byte, packetSize int, sampleRate int) []byte {
	absoluteStartTime := streamer.MediaClock.GetStartTimeInt64()
	playAt := streamer.MediaClock.GetSentTimeInt64()
	envelope := make([]byte, headerSize+packetSize)
	binary.BigEndian.PutUint64(envelope[0:8], uint64(absoluteStartTime))
	binary.BigEndian.PutUint64(envelope[8:16], uint64(playAt))
	binary.BigEndian.PutUint32(envelope[16:20], uint32(sampleRate))
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

func (streamer *AudioStreamer) StreamAudioToAllLoop(stop chan struct{}) {
	for {
		select {
		case <-stop:
			logging.Log(logPrefix, "Stopping Audio Streamer...")
			return
		default:
			streamer.mu.Lock()
			if streamer.TrackQueue.Len() == 0 {
				streamer.mu.Unlock()
				time.Sleep(streamer.SleepInterval)
				continue
			}
			streamer.MediaClock = *clock.NewMediaClock()
			streamer.MediaClock.UpdateStartTime(int(streamer.LookAhead))
			streamer.mu.Unlock()
			streamer.StreamAudioToAll(stop)
		}
	}
}
