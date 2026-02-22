package clock

import (
	"testing"
	"time"
)

func TestUpdateMediaTime(t *testing.T) {
	// arrange
	mediaClock := NewMediaClock()
	timePassed := time.Since(mediaClock.StartTime)
	// act
	mediaClock.UpdateMediaTime()

	// assert
	if mediaClock.MediaTime < timePassed {
		t.Fatal("UpdateMediaTime failed: Expected to be greater than " + timePassed.String() + " but was " + mediaClock.MediaTime.String())
	}
}

func TestStampTime(t *testing.T) {
	// arrange
	mediaClock := NewMediaClock()
	offset := 20 * time.Millisecond
	// act
	timeStamp := mediaClock.StampTime(offset)

	// assert
	expected := mediaClock.MediaTime + offset
	if timeStamp != expected {
		t.Fatal("UpdateMediaTime failed: Expected to be " + expected.String() + " but was " + timeStamp.String())
	}
}
