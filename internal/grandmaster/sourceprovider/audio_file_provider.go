package sourceprovider

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gopxl/beep/v2"
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

	if format.SampleRate != targetSampleRate {
		resampled := beep.Resample(4, format.SampleRate, targetSampleRate, streamer)
		decoder.Streamer = resampled
	}

	return decoder, nil
}

func (provider *AudioProvider) ListSongs() []string {
	files, _ := fs.Glob(provider.rootFs, "*.mp3")
	return files
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

func floatToInt16(f float64) int16 {
	if f < -1.0 {
		f = -1.0
	} else if f > 1.0 {
		f = 1.0
	}
	return int16(f * 32767)
}
