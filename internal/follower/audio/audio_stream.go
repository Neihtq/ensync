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
	return &AudioStream{
		isBuffering: true,
		clock:       clock,
	}
}

func (stream *AudioStream) WriteToBuffer(data []byte, playAt int64) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	tempBuffer := make([]byte, len(data))
	numBytes := copy(tempBuffer, data)

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

	if stream.chunks.Len() == 0 {
		zero(playBuffer)
		return len(playBuffer), nil
	}

	targetChunk := stream.chunks.Front()
	startTime := stream.clock.GetStartTime()
	if startTime.IsZero() {
		zero(playBuffer)
		return len(playBuffer), nil
	}

	targetPlayTime := startTime.Add(time.Duration(targetChunk.playAt))
	now := stream.clock.Now()
	drift := now.Sub(targetPlayTime)

	// too early --> silence
	if drift < 0 {
		zero(playBuffer)
		return len(playBuffer), nil
	}

	// too late --> drop chunk
	if drift > 20*time.Millisecond {
		stream.chunks.PopFront()
		stream.bufferSize -= len(targetChunk.data)
		return 0, nil
	}

	// play
	n := copy(playBuffer, targetChunk.data)
	if n < len(targetChunk.data) {
		stream.chunks.Set(0, AudioChunk{
			data:   targetChunk.data[n:],
			playAt: targetChunk.playAt,
		})
	} else {
		stream.chunks.PopFront()
	}

	stream.bufferSize -= n
	return n, nil
}

func zero(buffer []byte) {
	for i := range buffer {
		buffer[i] = 0
	}
}
