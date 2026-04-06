package sourceprovider

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gopxl/beep/v2/mp3"
)

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
	audioFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening audio source at %s: %w", path, err)
	}

	streamer, format, err := mp3.Decode(audioFile)
	if err != nil {
		log.Fatal(err)
	}
	decoder := &Decoder{
		Streamer:   streamer,
		Closer:     audioFile,
		SampleRate: sampleRate,
		Channels:   2,
	}

	Resample(format, streamer, decoder)
	return decoder, nil
}

func (provider *AudioProvider) ListSongs() []string {
	files, _ := fs.Glob(provider.rootFs, "*.mp3")
	return files
}
