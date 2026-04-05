// Package sourceprovider provides functions and interfaces to fetch audio sources
package sourceprovider

import (
	"io"
	"io/fs"
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
	ListSongs() []string
}

type AudioProvider struct {
	root fs.FS
}

func NewAudioProvider(root string) *AudioProvider {
	dir := os.DirFS(root)
	return &AudioProvider{root: dir}
}

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

func (provider *AudioProvider) ListSongs() []string {
	files, _ := fs.Glob(provider.root, "*.mp3")
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
