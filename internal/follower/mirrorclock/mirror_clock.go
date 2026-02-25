// Package mirrorclock implements the mirror clock that adjusts the local clock with the grandmasters clock by keeping an offset
package mirrorclock

import (
	"sync"
	"time"
)

type MirrorClock struct {
	mu          sync.Mutex
	VirtualTime time.Time
	Offset      time.Duration
	StartTime   time.Time
}

func NewMirrorClock() *MirrorClock {
	return &MirrorClock{
		VirtualTime: time.Now(),
		StartTime:   time.Time{},
	}
}

func (clock *MirrorClock) UpdateOffset(ts uint64) {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	timestamp := time.Unix(0, int64(ts))
	clock.Offset = time.Until(timestamp)
	clock.VirtualTime = time.Now().Add(clock.Offset)
}

func (clock *MirrorClock) Now() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	return time.Now().Add(clock.Offset)
}

func (clock *MirrorClock) ResetStartTime() {
	clock.StartTime = time.Time{}
}

func (clock *MirrorClock) InitStartTime(playAt int64) {
	clock.StartTime = clock.Now().Add(-time.Duration(playAt))
}

func (clock *MirrorClock) GetStartTimeInt64() int64 {
	return clock.StartTime.UnixNano()
}

func (clock *MirrorClock) GetTargetPlayTime(playAt int64) time.Time {
	return clock.StartTime.Add(time.Duration(playAt))
}
