package mirrorclock

import (
	"testing"
)

func TestMirrorClockNow(t *testing.T) {
	// arrange
	mirrorClock := NewMirrorClock()

	// act
	mirrorClock.Now()
}
