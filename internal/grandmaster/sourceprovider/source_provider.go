// Package sourceprovider provides functions and interfaces to fetch audio sources
package sourceprovider

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/go-mp3"
	resampling "github.com/tphakala/go-audio-resampler"
)



type Decoder struct {
	io.Reader
	io.Closer
	SampleRate int64
	Channels   int64
}

type SourceProvider interface {
	GetSource(trackIdentifier string) (*Decoder, error)
	ListSongs() []string
}

type AudioProvider struct {
	rootFs fs.FS
	root   string
}

func NewAudioProvider(root string) *AudioProvider {
	rootFs := os.DirFS(root)
	return &AudioProvider{
		rootFs: rootFs,
		root:   root,
	}
}

func (provider *AudioProvider) GetSource(trackIdentifier string) (*Decoder, error) {
	path := filepath.Join(provider.root, trackIdentifier)
	fmt.Println("Opening audio file", path)
	audioSource, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening audio source at %s: %w", path, err)
	}

	decoder, err := mp3.NewDecoder(audioSource)
	if err != nil {
		audioSource.Close()
		return nil, fmt.Errorf("error creating mp3 decoder: %w", err)
	}

	// Lock the output to 48000Hz using an advanced polyphase FIR resampler!
	var r resampling.Resampler
	if decoder.SampleRate() != 48000 {
		config := &resampling.Config{
			InputRate:  float64(decoder.SampleRate()),
			OutputRate: 48000,
			Channels:   2,
			Quality:    resampling.QualitySpec{Preset: resampling.QualityHigh},
		}
		r, err = resampling.New(config)
		if err != nil {
			audioSource.Close()
			return nil, fmt.Errorf("error creating resampler: %w", err)
		}
	}

	resampler := &ResamplingReader{
		decoder:    decoder,
		targetRate: 48000,
		resampler:  r,
	}

	return &Decoder{
		Reader:     resampler,
		Closer:     audioSource,
		SampleRate: 48000,
		Channels:   2,
	}, nil
}

func (provider *AudioProvider) ListSongs() []string {
	files, _ := fs.Glob(provider.rootFs, "*.mp3")
	return files
}

type MockSourceProvider struct{}

func (provider *MockSourceProvider) ListSongs() []string {
	return []string{"mock1", "mock2"}
}

func (provider *MockSourceProvider) GetSource(filePath string) (*Decoder, error) {
	reader := strings.NewReader(filePath)
	return &Decoder{
		Reader:     reader,
		Closer:     io.NopCloser(reader),
		SampleRate: 48000,
		Channels:   2,
	}, nil
}
