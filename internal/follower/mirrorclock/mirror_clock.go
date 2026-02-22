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
}

func NewMirrorClock() *MirrorClock {
	return &MirrorClock{VirtualTime: time.Now()}
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
