// Package clock implements the virtual clock of the grandmaster
// This dictates the source of truth time for followers to be in sync
package clock

import "time"

type MediaClock struct {
	StartTime time.Time
	MediaTime time.Duration
	SentTime  time.Duration
	IsPlaying bool
}

func NewMediaClock() *MediaClock {
	return &MediaClock{
		StartTime: time.Now(),
		MediaTime: 0,
		SentTime:  0,
		IsPlaying: false,
	}
}

func (clock *MediaClock) UpdateStartTime() {
	clock.StartTime = time.Now()
}

func (clock *MediaClock) GetSentTimeInt64() int64 {
	return clock.SentTime.Nanoseconds()
}

func (clock *MediaClock) AddToSentTime(durationSent int64) {
	clock.SentTime += time.Duration(durationSent)
}

func (clock *MediaClock) GetMediaTimeInt64() int64 {
	return clock.MediaTime.Nanoseconds()
}

func (clock *MediaClock) UpdateMediaTime() {
	clock.MediaTime = time.Since(clock.StartTime)
}

func (clock *MediaClock) StampTime(offset int64) int64 {
	return time.Now().UnixNano() + offset
}
