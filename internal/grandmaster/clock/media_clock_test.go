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
	offset := 200 * time.Millisecond
	startTime := time.Now().UnixNano()
	// act
	timeStamp := mediaClock.StampTime(offset.Nanoseconds())

	// assert
	expected := startTime + offset.Nanoseconds()
	if timeStamp < expected {
		t.Fatalf("UpdateMediaTime failed: Expected to be greater than %d but was %d", expected, timeStamp)
	}
}
