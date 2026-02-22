package mirrorclock

import (
	"testing"
	"time"
)

func TestMirrorClockUpdate(t *testing.T) {
	// arrange
	mirrorClock := NewMirrorClock()
	virtualTime := mirrorClock.VirtualTime
	offset := mirrorClock.Offset

	// act
	time.Sleep(10 * time.Millisecond)
	mirrorClock.UpdateOffset(uint64(time.Now().UnixNano()))

	// assert
	if mirrorClock.VirtualTime.Equal(virtualTime) {
		t.Fatalf("MirrorClock update failed: virtual time should be different from %s but was the same", virtualTime.String())
	}
	if mirrorClock.Offset == offset {
		t.Fatalf("MirrorClock update failed: Offset should be different from %s but was the same", offset.String())
	}
}

func TestMirrorClockNow(t *testing.T) {
	// arrange
	mirrorClock := NewMirrorClock()

	// act
	mirrorClock.Now()
}
