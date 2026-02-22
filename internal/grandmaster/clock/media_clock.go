// Package clock implements the virtual clock of the grandmaster
// This dictates the source of truth time for followers to be in sync
package clock

import "time"

type MediaClock struct {
	StartTime time.Time
	MediaTime time.Duration
}

func NewMediaClock() *MediaClock {
	return &MediaClock{StartTime: time.Now()}
}

func (clock *MediaClock) UpdateMediaTime() {
	clock.MediaTime = time.Since(clock.StartTime)
}

func (clock *MediaClock) StampTime(offset time.Duration) time.Duration {
	return clock.MediaTime + offset
}
