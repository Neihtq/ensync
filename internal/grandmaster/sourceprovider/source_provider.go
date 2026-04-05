// Package sourceprovider provides functions and interfaces to fetch audio sources
package sourceprovider

import (
	"io"
	"os"
	"strings"

	"github.com/hajimehoshi/go-mp3"
)

type Decoder struct {
	io.Reader
	io.Closer
	SampleRate int64
	Channels   int64
}

type SourceProvider interface {
	GetSource(filePath string) *Decoder
}

type AudioProvider struct{}

func (provider *AudioProvider) GetSource(filePath string) *Decoder {
	audioSource, err := os.Open(filePath)
	if err != nil {
		panic("Error opening audio source at " + filePath)
	}

	decoder, err := mp3.NewDecoder(audioSource)
	if err != nil {
		panic("Error creating mp3 decoder " + err.Error())
	}

	return &Decoder{
		Reader:     decoder,
		Closer:     audioSource,
		SampleRate: int64(decoder.SampleRate()),
		Channels:   2,
	}
}

type MockSourceProvider struct{}

func (provider *MockSourceProvider) GetSource(filePath string) *Decoder {
	reader := strings.NewReader(filePath)
	return &Decoder{
		Reader:     reader,
		Closer:     io.NopCloser(reader),
		SampleRate: 48000,
		Channels:   2,
	}
}
