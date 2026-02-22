package audiostreamer

import (
	"io"
	"os"
	"strings"

	"github.com/hajimehoshi/go-mp3"
)

type SourceProvider interface {
	GetSource(filePath string) io.ReadCloser
}

type AudioProvider struct{}

func (provider *AudioProvider) GetSource(filePath string) io.ReadCloser {
	audioSource, err := os.Open(filePath)
	if err != nil {
		panic("Error opening audio source at " + filePath)
	}

	decoder, err := mp3.NewDecoder(audioSource)
	if err != nil {
		panic("Error creating mp3 decoder " + err.Error())
	}

	return &struct {
		io.Reader
		io.Closer
	}{decoder, audioSource}
}

type MockSourceProvider struct{}

func (provider *MockSourceProvider) GetSource(filePath string) io.ReadCloser {
	reader := strings.NewReader(filePath)
	return &struct {
		io.Reader
		io.Closer
	}{reader, io.NopCloser(reader)}
}
