// Package sourceprovider provides functions and interfaces to fetch audio sources
package sourceprovider

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
	GetSource(trackIdentifier string) *Decoder
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

func (provider *AudioProvider) GetSource(trackIdentifier string) *Decoder {
	path := filepath.Join(provider.root, trackIdentifier)
	audioSource, err := os.Open(path)
	if err != nil {
		panic("Error opening audio source at " + path)
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

func (provider *AudioProvider) ListSongs() []string {
	files, _ := fs.Glob(provider.rootFs, "*.mp3")
	return files
}

type MockSourceProvider struct{}

func (provider *MockSourceProvider) ListSongs() []string {
	return []string{"mock1", "mock2"}
}

func (provider *MockSourceProvider) GetSource(filePath string) *Decoder {
	reader := strings.NewReader(filePath)
	return &Decoder{
		Reader:     reader,
		Closer:     io.NopCloser(reader),
		SampleRate: 48000,
		Channels:   2,
	}
}
