package audio

import (
	"sync"
	"time"

	"ensync/internal/follower/mirrorclock"

	"github.com/gammazero/deque"
)

type AudioChunk struct {
	data   []byte
	playAt int64
}

type AudioStream struct {
	mu          sync.Mutex
	chunks      deque.Deque[AudioChunk]
	bufferSize  int
	isBuffering bool
	clock       *mirrorclock.MirrorClock
	lastPlayAt  int64
}

func NewAudioStream(clock *mirrorclock.MirrorClock) *AudioStream {
	return &AudioStream{isBuffering: true, bufferSize: 0, clock: clock}
}

func (stream *AudioStream) WriteToBuffer(data []byte, playAt int64) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	tempBuffer := make([]byte, len(data))
	numBytes := copy(tempBuffer, data)

	if playAt < stream.lastPlayAt || (playAt == 0 && stream.lastPlayAt == 0) {
		stream.clock.ResetStartTime()
		stream.chunks.Clear()
		stream.bufferSize = 0
		stream.isBuffering = false
	}

	stream.chunks.PushBack(AudioChunk{
		data:   tempBuffer,
		playAt: playAt,
	})
	stream.bufferSize += numBytes
}

func (stream *AudioStream) Read(playBuffer []byte) (int, error) {
	const threshold = 192000
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.isBuffering {
		if stream.bufferSize < threshold {
			for i := range playBuffer {
				playBuffer[i] = 0
			}
			return len(playBuffer), nil
		}
		stream.isBuffering = false
	}

	if stream.bufferSize == 0 {
		stream.isBuffering = true
		for i := range playBuffer {
			playBuffer[i] = 0
		}
	}

	targetChunk := stream.chunks.Front()
	playAt := targetChunk.playAt

	if stream.clock.StartTime.IsZero() {
		stream.clock.InitStartTime(playAt)
	}
	stream.lastPlayAt = playAt
	tolerance := 10 * time.Millisecond
	targetPlayTime := stream.clock.GetTargetPlayTime(playAt)
	if stream.clock.Now().Add(tolerance).Before(targetPlayTime) {
		for i := range playBuffer {
			playBuffer[i] = 0
		}
		return len(playBuffer), nil
	}

	numBytes := copy(playBuffer, targetChunk.data)

	if numBytes < len(targetChunk.data) {
		stream.chunks.Set(0, AudioChunk{
			data:   targetChunk.data[numBytes:],
			playAt: targetChunk.playAt,
		})
	} else {
		stream.chunks.PopFront()
	}
	stream.bufferSize -= numBytes

	return numBytes, nil
}
