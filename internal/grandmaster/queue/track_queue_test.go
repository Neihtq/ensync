package queue

import (
	"sync"
	"testing"
)

func TestTrackQueue_PushAndPop(t *testing.T) {
	q := NewTrackQueue()

	if q.Len() != 0 {
		t.Errorf("expected length 0, got %d", q.Len())
	}

	q.PushBack("track1.mp3")
	q.PushBack("track2.mp3")

	if q.Len() != 2 {
		t.Errorf("expected length 2, got %d", q.Len())
	}

	item := q.PopFront()
	if item != "track1.mp3" {
		t.Errorf("expected 'track1.mp3', got '%s'", item)
	}

	if q.Len() != 1 {
		t.Errorf("expected length 1, got %d", q.Len())
	}

	item = q.PopFront()
	if item != "track2.mp3" {
		t.Errorf("expected 'track2.mp3', got '%s'", item)
	}

	if q.Len() != 0 {
		t.Errorf("expected length 0, got %d", q.Len())
	}
}

func TestTrackQueue_Concurrency(t *testing.T) {
	q := NewTrackQueue()
	var wg sync.WaitGroup

	numGoroutines := 50
	numItems := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numItems; j++ {
				q.PushBack("track.mp3")
			}
		}()
	}

	wg.Wait()

	expectedLen := numGoroutines * numItems
	if q.Len() != expectedLen {
		t.Errorf("expected length %d after concurrent pushes, got %d", expectedLen, q.Len())
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numItems; j++ {
				q.PopFront()
			}
		}()
	}

	wg.Wait()

	if q.Len() != 0 {
		t.Errorf("expected length 0 after concurrent pops, got %d", q.Len())
	}
}

func TestTrackQueue_GetAndSetNowPlaying(t *testing.T) {
	q := NewTrackQueue()

	if q.NowPlaying != "" {
		t.Errorf("expected NowPlaying to be empty string, got %s", q.NowPlaying)
	}
	trackID := "SomeTrack.mp3"
	q.SetNowPlaying(trackID)
	if q.NowPlaying != trackID {
		t.Errorf("expected NowPlaying to be '%s', got %s", trackID, q.NowPlaying)
	}
	if q.GetNowPlaying() != trackID {
		t.Errorf("expected NowPlaying to be '%s', got %s", trackID, q.GetNowPlaying())
	}
}
