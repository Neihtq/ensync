package audio

import "sync"

type AudioStream struct {
	mu          sync.Mutex
	data        []byte
	isBuffering bool
}

func NewAudioStream() *AudioStream {
	return &AudioStream{isBuffering: true}
}

func (stream *AudioStream) WriteToBuffer(data []byte) {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	stream.data = append(stream.data, data...)
}

func (stream *AudioStream) Read(playBuffer []byte) (int, error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	const threshold = 38400
	if stream.isBuffering {
		if len(stream.data) < threshold {
			return 0, nil
		}
		stream.isBuffering = false
	}

	if len(stream.data) == 0 {
		stream.isBuffering = true
		return 0, nil
	}

	numBytes := copy(playBuffer, stream.data)
	stream.data = stream.data[numBytes:]

	return numBytes, nil
}
