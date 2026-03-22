package audio

import (
	"fmt"
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
	mu               sync.Mutex
	chunks           deque.Deque[AudioChunk]
	bufferSize       int
	isBuffering      bool
	clock            *mirrorclock.MirrorClock
	playbackDelay    time.Duration
	hasAligned       bool
	alignmentSamples []time.Duration
}

func NewAudioStream(clock *mirrorclock.MirrorClock) *AudioStream {
	return &AudioStream{
		isBuffering:   true,
		clock:         clock,
		playbackDelay: 100 * time.Millisecond,
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

	// if !stream.hasAligned {
	// 	stream.alignDelayWithCurrentTime(startTime, targetChunk)
	// }

	startPlaybackTime := startTime.Add(time.Duration(targetChunk.playAt))
	now := stream.clock.Now()
	fmt.Printf("\rTime: %v ,\nplayback time: %v", now, startPlaybackTime)
	if now.Before(startPlaybackTime) {
		zero(playBuffer)
		return len(playBuffer), nil
	}

	// clockDrift := stream.calcClockDrift(startTime, targetChunk)
	// if !stream.validateClockDrift(playBuffer, clockDrift, targetChunk) {
	// 	return len(playBuffer), nil
	// }

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

func (stream *AudioStream) calcClockDrift(startTime time.Time, targetChunk AudioChunk) time.Duration {
	targetPlayTime := startTime.Add(time.Duration(targetChunk.playAt)).Add(stream.playbackDelay)
	now := stream.clock.Now()
	drift := now.Sub(targetPlayTime)

	return drift
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
		return false
	}

	return true
}

func (stream *AudioStream) alignDelayWithCurrentTime(startTime time.Time, targetChunk AudioChunk) {
	targetPlayTime := startTime.Add(time.Duration(targetChunk.playAt))
	now := stream.clock.Now()

	currentOffset := now.Sub(targetPlayTime)
	stream.alignmentSamples = append(stream.alignmentSamples, currentOffset)

	const alignmentSampleThreshold = 10
	if len(stream.alignmentSamples) >= alignmentSampleThreshold {
		var total time.Duration
		for _, sample := range stream.alignmentSamples {
			total += sample
		}
		avgDrift := total / alignmentSampleThreshold
		stream.playbackDelay += avgDrift

		stream.hasAligned = true
		stream.alignmentSamples = nil
	}
}

func zero(buffer []byte) {
	for i := range buffer {
		buffer[i] = 0
	}
}
