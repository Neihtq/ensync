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

func TestTrackQueue_Last(t *testing.T) {
	q := NewTrackQueue()
	q.PushBack("first.mp3")
	q.PushBack("second.mp3")
	q.PushBack("third.mp3")

	last := q.Last()
	if last != "third.mp3" {
		t.Errorf("expected 'third.mp3', got '%s'", last)
	}
}

func TestTrackQueue_GetAllItems(t *testing.T) {
	q := NewTrackQueue()
	q.PushBack("a.mp3")
	q.PushBack("b.mp3")
	q.PushBack("c.mp3")

	items := q.GetAllItems()
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0] != "a.mp3" || items[1] != "b.mp3" || items[2] != "c.mp3" {
		t.Errorf("unexpected items: %v", items)
	}
	if q.Len() != 3 {
		t.Errorf("expected queue length still 3, got %d", q.Len())
	}
}

func TestTrackQueue_GetAllItems_Empty(t *testing.T) {
	q := NewTrackQueue()
	items := q.GetAllItems()
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestTrackQueue_CallHookFiresOnPush(t *testing.T) {
	q := NewTrackQueue()
	hookCalled := false
	var hookNowPlaying string
	var hookItems []string

	q.SetCallbackHook(func(nowPlaying string, queueItems []string) {
		hookCalled = true
		hookNowPlaying = nowPlaying
		hookItems = queueItems
	})

	q.SetNowPlaying("playing.mp3")
	q.PushBack("queued.mp3")

	if !hookCalled {
		t.Fatal("expected hook to be called on PushBack")
	}
	if hookNowPlaying != "playing.mp3" {
		t.Errorf("expected nowPlaying 'playing.mp3', got '%s'", hookNowPlaying)
	}
	if len(hookItems) != 1 || hookItems[0] != "queued.mp3" {
		t.Errorf("expected hook items ['queued.mp3'], got %v", hookItems)
	}
}

func TestTrackQueue_CallHookFiresOnPop(t *testing.T) {
	q := NewTrackQueue()
	q.PushBack("track.mp3")
	hookCalled := false
	q.SetCallbackHook(func(nowPlaying string, queueItems []string) {
		hookCalled = true
	})
	q.PopFront()
	if !hookCalled {
		t.Fatal("expected hook to be called on PopFront")
	}
}

func TestTrackQueue_CallHookWithoutHookSet(t *testing.T) {
	q := NewTrackQueue()
	// Should not panic when no hook is set
	q.PushBack("track.mp3")
	q.PopFront()
	q.SetNowPlaying("test")
	q.CallHook()
}
