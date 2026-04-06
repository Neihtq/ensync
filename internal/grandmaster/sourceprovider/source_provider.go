// Package sourceprovider provides functions and interfaces to fetch audio sources
package sourceprovider

import (
	"io"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/generators"
)

const (
	sampleRate       = 48000
	targetSampleRate = beep.SampleRate(sampleRate)
)

type Decoder struct {
	Streamer   beep.Streamer
	Closer     io.Closer
	SampleRate int64
	Channels   int64
}

type SourceProvider interface {
	GetSource(trackIdentifier string) (*Decoder, error)
	ListSongs() []string
}

type MockSourceProvider struct{}

func (provider *MockSourceProvider) ListSongs() []string {
	return []string{"mock1", "mock2"}
}

func (provider *MockSourceProvider) GetSource(filePath string) (*Decoder, error) {
	reader := strings.NewReader(filePath)
	dummyStreamer := generators.Silence(12000)
	return &Decoder{
		Streamer:   dummyStreamer,
		Closer:     io.NopCloser(reader),
		SampleRate: sampleRate,
		Channels:   2,
	}, nil
}
