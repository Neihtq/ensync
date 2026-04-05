// Package queue implements a track queue used as shared object between between services
package queue

import (
	"sync"

	"github.com/gammazero/deque"
)

type TrackQueue struct {
	mu sync.Mutex

	queue *deque.Deque[string]
}

func NewTrackQueue() *TrackQueue {
	return &TrackQueue{queue: &deque.Deque[string]{}}
}

func (tq *TrackQueue) PushBack(item string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.queue.PushBack(item)
}

func (tq *TrackQueue) PopFront() string {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return tq.queue.PopFront()
}

func (tq *TrackQueue) Len() int {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return tq.queue.Len()
}
