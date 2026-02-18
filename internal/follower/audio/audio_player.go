// Package audio implements audio play functionalities
package audio

import (
	"io"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

func PlayAudio(source io.Reader) {
	decodedMp3, err := mp3.NewDecoder(source)
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}

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

	player := otoContext.NewPlayer(decodedMp3)
	player.Play()

	for player.IsPlaying() {
		time.Sleep(time.Millisecond * 100)
	}
}
