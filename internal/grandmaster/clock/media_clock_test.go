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

func TestUpdateStartTime(t *testing.T) {
	// arrange
	mediaClock := NewMediaClock()
	lookAhead := 2

	// act
	mediaClock.UpdateStartTime(lookAhead)

	// assert
	currTime := time.Now()
	threshold := 2 * time.Second
	diff := mediaClock.StartTime.Sub(currTime)
	if diff > threshold {
		t.Fatalf("UpdateStartTime failed: Expected diff to be less than %d but was %d (timeNow + lookAhead)", threshold, diff)
	}
}

func TestGetSentTimeInt64(t *testing.T) {
	// arrange
	mediaClock := NewMediaClock()

	// act
	sentTime := mediaClock.GetSentTimeInt64()

	// assert
	if sentTime != 0 {
		t.Fatalf("GetSentTimeInt64 failed: Expected 0 but received %d", sentTime)
	}
}

func TestAddToSentTime(t *testing.T) {
	// arrange
	mediaClock := NewMediaClock()

	// act
	mediaClock.AddToSentTime(10)

	// assert
	sentTime := mediaClock.GetSentTimeInt64()
	if sentTime != 10 {
		t.Fatalf("AddToSentTime failed: Expected result to be 10 but received %d", sentTime)
	}
}

func TestMediaTimeInt64(t *testing.T) {
	// arrange
	mediaClock := NewMediaClock()

	// act
	mediaTime := mediaClock.GetMediaTimeInt64()

	// assert
	if mediaTime != 0 {
		t.Fatalf("GetMediaTime64 failed: Expected 0 but received %d", mediaTime)
	}
}
