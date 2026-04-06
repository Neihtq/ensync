package sourceprovider

import "github.com/gopxl/beep/v2"

const (
	sampleRate       = 48000
	targetSampleRate = beep.SampleRate(sampleRate)
)

func Resample(format beep.Format, streamer beep.Streamer, decoder *Decoder) {
	if format.SampleRate != targetSampleRate {
		resampled := beep.Resample(4, format.SampleRate, targetSampleRate, streamer)
		decoder.Streamer = resampled
	}
}

func floatToInt16(f float64) int16 {
	if f < -1.0 {
		f = -1.0
	} else if f > 1.0 {
		f = 1.0
	}
	return int16(f * 32767)
}
