// Package queue implements a track queue used as shared object between between services
package queue

import (
	"sync"

	"github.com/gammazero/deque"
)

type TrackQueue struct {
	mu    sync.Mutex
	queue *deque.Deque[string]

	NowPlaying     string
	OnQueueChanged func(nowPlaying string, queueItems []string)
}

func NewTrackQueue() *TrackQueue {
	return &TrackQueue{queue: &deque.Deque[string]{}}
}

func (tq *TrackQueue) PushBack(item string) {
	tq.mu.Lock()
	tq.queue.PushBack(item)
	tq.mu.Unlock()

	tq.CallHook()
}

func (tq *TrackQueue) PopFront() string {
	tq.mu.Lock()
	track := tq.queue.PopFront()
	tq.mu.Unlock()

	tq.CallHook()
	return track
}

func (tq *TrackQueue) Len() int {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return tq.queue.Len()
}

func (tq *TrackQueue) Last() string {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	return tq.queue.Back()
}

func (tq *TrackQueue) GetAllItems() []string {
	queueLength := tq.Len()

	tq.mu.Lock()
	defer tq.mu.Unlock()

	playList := make([]string, queueLength)
	for i := range queueLength {
		playList[i] = tq.queue.At(i)
	}

	return playList
}

func (tq *TrackQueue) SetNowPlaying(trackID string) {
	tq.mu.Lock()
	tq.NowPlaying = trackID
	tq.mu.Unlock()

	tq.CallHook()
}

func (tq *TrackQueue) GetNowPlaying() string {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	return tq.NowPlaying
}

func (tq *TrackQueue) SetCallbackHook(
	onQueueChanged func(nowPlaying string, queueItems []string),
) {
	tq.OnQueueChanged = onQueueChanged
}

func (tq *TrackQueue) CallHook() {
	if tq.OnQueueChanged != nil {
		nowPlaying := tq.GetNowPlaying()
		queueItems := tq.GetAllItems()
		tq.OnQueueChanged(nowPlaying, queueItems)
	}
}
