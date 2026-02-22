// Package audio implements audio play functionalities
package audio

import (
	"github.com/ebitengine/oto/v3"
)

type ErrorReporter interface {
	Err() error
}

type MockErrorReporter struct {
	mockErr error
}

func (er *MockErrorReporter) Err() error {
	return er.mockErr
}

func NewPlayer(audioStream *AudioStream) (*oto.Context, *oto.Player) {
	op := &oto.NewContextOptions{
		SampleRate:   48000,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	}
	otoContext, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan

	player := otoContext.NewPlayer(audioStream)
	player.Play()

	return otoContext, player
}
