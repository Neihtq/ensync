package sourceprovider

import (
	"io"

	"github.com/hajimehoshi/go-mp3"
	resampling "github.com/tphakala/go-audio-resampler"
)

// ResamplingReader forces playback to a fixed sample rate.
type ResamplingReader struct {
	decoder    *mp3.Decoder
	targetRate int
	resampler  resampling.Resampler
	outBuffer  []byte
}

func bytesToFloat64Interleaved(b []byte) []float64 {
	f := make([]float64, len(b)/2)
	for i := 0; i < len(b)/2; i++ {
		val := int16(b[i*2]) | int16(b[i*2+1])<<8
		f[i] = float64(val) / 32768.0
	}
	return f
}

func float64ToBytesInterleaved(f []float64) []byte {
	b := make([]byte, len(f)*2)
	for i, val := range f {
		if val > 1.0 {
			val = 1.0
		}
		if val < -1.0 {
			val = -1.0
		}
		intVal := int16(val * 32767.0)
		b[i*2] = byte(intVal)
		b[i*2+1] = byte(intVal >> 8)
	}
	return b
}

func (r *ResamplingReader) Read(p []byte) (n int, err error) {
	if r.resampler == nil {
		return r.decoder.Read(p)
	}

	for len(r.outBuffer) < len(p) {
		inFramesRequired := int(int64(len(p)/4) * int64(r.decoder.SampleRate()) / int64(r.targetRate))
		if inFramesRequired < 1024 {
			inFramesRequired = 1024
		}
		inBuf := make([]byte, inFramesRequired*4)
		inN, inErr := io.ReadFull(r.decoder, inBuf)

		if inN > 0 {
			f := bytesToFloat64Interleaved(inBuf[:inN])
			outF, errProcess := r.resampler.Process(f)
			if errProcess != nil {
				return 0, errProcess
			}
			outB := float64ToBytesInterleaved(outF)
			r.outBuffer = append(r.outBuffer, outB...)
		}

		if inErr != nil {
			if inErr == io.EOF || inErr == io.ErrUnexpectedEOF {
				outF, _ := r.resampler.Flush()
				if len(outF) > 0 {
					outB := float64ToBytesInterleaved(outF)
					r.outBuffer = append(r.outBuffer, outB...)
				}
				if len(r.outBuffer) == 0 {
					return 0, io.EOF
				}
				break
			}
			return 0, inErr
		}
	}

	n = copy(p, r.outBuffer)
	r.outBuffer = r.outBuffer[n:]
	return n, nil
}
