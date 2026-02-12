// Package clock implements the virtual clock of the grandmaster
// This dictates the source of truth time for followers to be in sync
package clock

import "time"

type Clock struct {
	StartTime time.Time
}

func NewClock() *Clock {
	return &Clock{StartTime: time.Now()}
}

func (clock *Clock) GetTimePassed() time.Duration {
	return time.Since(clock.StartTime)
}
