// Package sourceprovider provides functions and interfaces to fetch audio sources
package sourceprovider

import (
	"encoding/binary"
	"io"
	"strings"

	"ensync/internal/grandmaster/navidrome"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/generators"
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
	SearchSong(query string) []navidrome.Song
	GetTitle(trackIdentifier string) string
}

func (d *Decoder) Close() error {
	return d.Closer.Close()
}

func (d *Decoder) Read(p []byte) (n int, err error) {
	numSamples := len(p) / 4
	samples := make([][2]float64, numSamples)

	sn, ok := d.Streamer.Stream(samples)
	if !ok && sn == 0 {
		return 0, io.EOF
	}

	for i := range sn {
		leftInt := floatToInt16(samples[i][0])
		binary.LittleEndian.PutUint16(p[i*4:], uint16(leftInt))

		rightInt := floatToInt16(samples[i][1])
		binary.LittleEndian.PutUint16(p[i*4+2:], uint16(rightInt))
	}

	return sn * 4, nil
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

func (provider *MockSourceProvider) SearchSong(query string) []navidrome.Song {
	if query == "oasis" {
		return []navidrome.Song{
			{ID: "oasis1", Title: "Wonderwall", Artist: "Oasis"},
		}
	}
	return []navidrome.Song{}
}

func (provider *MockSourceProvider) GetTitle(trackIdentifier string) string {
	return trackIdentifier
}
