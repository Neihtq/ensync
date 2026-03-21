// Package mirrorclock implements the mirror clock that adjusts the local clock with the grandmasters clock by keeping an offset
package mirrorclock

import (
	"sync"
	"time"
)

type MirrorClock struct {
	mu        sync.Mutex
	StartTime time.Time
	Offset    float64
}

func NewMirrorClock() *MirrorClock {
	return &MirrorClock{
		StartTime: time.Time{},
	}
}

func (clock *MirrorClock) Now() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	return time.Unix(0, time.Now().UnixNano()+int64(clock.Offset))
}

func (clock *MirrorClock) ResetStartTime() {
	clock.StartTime = time.Time{}
}

func (clock *MirrorClock) InitStartTime(targetTime int64) {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	clock.StartTime = time.Unix(0, targetTime)
}

func (clock *MirrorClock) GetStartTimeInt64() int64 {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.StartTime.UnixNano()
}

func (clock *MirrorClock) GetTargetPlayTime(playAt int64) time.Time {
	return clock.StartTime.Add(time.Duration(playAt))
}

func (clock *MirrorClock) SyncTime(timeStamps []int64) {
	localSendTime, serverReceiveTime, serverSendTime, localReceiveTime := timeStamps[0], timeStamps[1], timeStamps[2], timeStamps[3]
	offset := ((serverReceiveTime - localSendTime) + (serverSendTime - localReceiveTime)) / 2
	delay := (localReceiveTime - localSendTime) - (serverSendTime - serverReceiveTime)
	clock.UpdateOffset(offset, delay)
}

func (clock *MirrorClock) UpdateOffset(measuredOffset int64, delay int64) {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	if delay > 50_000_000 {
		return
	}

	alpha := 0.02
	clock.Offset = clock.Offset*(1-alpha) + float64(measuredOffset)*alpha
}

func (clock *MirrorClock) GetOffset() float64 {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.Offset
}

func (clock *MirrorClock) GetStartTime() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	return clock.StartTime
}
