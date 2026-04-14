package audio

import (
	"sync"
	"time"

	"ensync/internal/follower/mirrorclock"

	"github.com/gammazero/deque"
)

const startupBytes = 150_000

type AudioChunk struct {
	data   []byte
	playAt int64
}

type AudioStream struct {
	mu                sync.Mutex
	chunks            deque.Deque[AudioChunk]
	bufferSize        int
	isBuffering       bool
	clock             *mirrorclock.MirrorClock
	playbackDelay     time.Duration
	hasAligned        bool
	alignmentSamples  []time.Duration
	hasStartedPlaying bool
	sampleRate        int
}

func (stream *AudioStream) SetSampleRate(sr int) {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	stream.sampleRate = sr
}

func (stream *AudioStream) GetSampleRate() int {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	return stream.sampleRate
}

func NewAudioStream(clock *mirrorclock.MirrorClock) *AudioStream {
	return &AudioStream{
		isBuffering:   true,
		clock:         clock,
		playbackDelay: 2000 * time.Millisecond,
	}
}

func (stream *AudioStream) WriteToBuffer(data []byte, playAt int64) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	var popped []AudioChunk
	for stream.chunks.Len() > 0 && stream.chunks.Back().playAt >= playAt {
		back := stream.chunks.Back()
		if back.playAt == playAt {
			for i := len(popped) - 1; i >= 0; i-- {
				stream.chunks.PushBack(popped[i])
			}
			return
		}
		popped = append(popped, stream.chunks.PopBack())
	}

	tempBuffer := make([]byte, len(data))
	numBytes := copy(tempBuffer, data)
	newChunk := AudioChunk{
		data:   tempBuffer,
		playAt: playAt,
	}

	stream.chunks.PushBack(newChunk)
	stream.bufferSize += numBytes

	// keep monotonic increasing queue
	for i := len(popped) - 1; i >= 0; i-- {
		stream.chunks.PushBack(popped[i])
	}
}

func (stream *AudioStream) Read(playBuffer []byte) (int, error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if !stream.bufferIsReady() {
		zero(playBuffer)
		return len(playBuffer), nil
	}

	targetChunk := stream.chunks.Front()
	startTime := stream.clock.GetStartTime()

	if startTime.IsZero() {
		stream.isBuffering = true
		zero(playBuffer)
		return len(playBuffer), nil
	}

	startPlaybackTime := startTime.Add(time.Duration(targetChunk.playAt))
	now := stream.clock.Now()

	if now.Before(startPlaybackTime) {
		durationToSilence := startPlaybackTime.Sub(now)

		if stream.hasStartedPlaying && durationToSilence < 10*time.Millisecond {
			stream.hasStartedPlaying = true
			return stream.playAudio(playBuffer, targetChunk), nil
		}

		bytesPerSec := 192000 // assuming 48000 Hz

		bytesToSilence := int(durationToSilence.Nanoseconds() * int64(bytesPerSec) / 1e9)
		bytesToSilence = (bytesToSilence / 4) * 4

		if bytesToSilence >= len(playBuffer) {
			zero(playBuffer)
			return len(playBuffer), nil
		} else if bytesToSilence > 0 {
			zero(playBuffer[:bytesToSilence])
			written := stream.playAudio(playBuffer[bytesToSilence:], targetChunk)
			stream.hasStartedPlaying = true
			return bytesToSilence + written, nil
		}
	}
	stream.hasStartedPlaying = true

	return stream.playAudio(playBuffer, targetChunk), nil
}

func (stream *AudioStream) playAudio(playBuffer []byte, targetChunk AudioChunk) int {
	playingBytes := copy(playBuffer, targetChunk.data)
	if playingBytes < len(targetChunk.data) {
		stream.chunks.Set(0, AudioChunk{
			data:   targetChunk.data[playingBytes:],
			playAt: targetChunk.playAt,
		})
	} else {
		stream.chunks.PopFront()
	}

	stream.bufferSize -= playingBytes
	return playingBytes
}

func (stream *AudioStream) validateClockDrift(playBuffer []byte, clockDrift time.Duration, targetChunk AudioChunk) bool {
	if clockDrift < -20*time.Millisecond {
		zero(playBuffer)
		return false
	}

	if clockDrift > 500*time.Millisecond {
		stream.chunks.PopFront()
		stream.bufferSize -= len(targetChunk.data)
		return false
	}

	return true
}

func (stream *AudioStream) bufferIsReady() bool {
	if stream.isBuffering {
		if stream.bufferSize < startupBytes {
			return false
		}
		stream.isBuffering = false
		stream.hasAligned = false
	}

	if stream.chunks.Len() == 0 {
		stream.isBuffering = true
		stream.hasStartedPlaying = false
		return false
	}

	return true
}

func zero(buffer []byte) {
	for i := range buffer {
		buffer[i] = 0
	}
}
