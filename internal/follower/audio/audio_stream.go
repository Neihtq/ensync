package audio

import (
	"sync"

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
}

func NewAudioStream() *AudioStream {
	return &AudioStream{isBuffering: true, bufferSize: 0}
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
	stream.mu.Lock()
	defer stream.mu.Unlock()

	const threshold = 38400
	if stream.isBuffering {
		if stream.bufferSize < threshold {
			return 0, nil
		}
		stream.isBuffering = false
	}

	if stream.bufferSize == 0 {
		stream.isBuffering = true
		return 0, nil
	}

	targetChunk := stream.chunks.Front()
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
