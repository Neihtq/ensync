package webservice

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/navidrome"
	"ensync/internal/grandmaster/queue"
	"ensync/internal/grandmaster/sourceprovider"
)

func TestNewWebServer(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	port := ":9999"

	server := NewWebServer(port, "", provider, registry, trackQueue)
	if server.Port != port {
		t.Errorf("Expected port %s, got %s", port, server.Port)
	}
	if server.SourceProvider != provider {
		t.Errorf("Expected source provider to be set properly")
	}
	if server.FollowersRegistry != registry {
		t.Errorf("Expected followers registry to be set properly")
	}
}

func TestGetSongs(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":9999", "", provider, registry, trackQueue)

	req := httptest.NewRequest(http.MethodGet, "/songs", nil)
	rr := httptest.NewRecorder()

	server.GetSongs(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string][]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	songs, ok := response["songs"]
	if !ok {
		t.Fatalf("Expected 'songs' key in response")
	}

	if len(songs) != 0 {
		t.Errorf("Unexpected songs list: %v", songs)
	}
}

func TestGetSongs_WithSearch(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":9999", "", provider, registry, trackQueue)

	req := httptest.NewRequest(http.MethodGet, "/songs?query=oasis", nil)
	rr := httptest.NewRecorder()

	server.GetSongs(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response struct {
		Songs []navidrome.Song `json:"songs"`
	}
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Songs) != 1 {
		t.Fatalf("Expected 1 song in search result, got %d", len(response.Songs))
	}

	if response.Songs[0].Title != "Wonderwall" {
		t.Errorf("Expected song title 'Wonderwall', got %s", response.Songs[0].Title)
	}
}

func TestListFollowers(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()

	registry.RegisterFollower("192.168.1.10", "8000")
	registry.RegisterFollower("192.168.1.11", "8000")

	server := NewWebServer(":9999", "", provider, registry, trackQueue)

	req := httptest.NewRequest(http.MethodGet, "/followers", nil)
	rr := httptest.NewRecorder()

	server.ListFollowers(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string][]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	followers, ok := response["followerUrls"]
	if !ok {
		t.Fatalf("Expected 'followerUrls' key in response")
	}

	// GetAllFollowers uses a map under the hood, so order is not guaranteed.
	if len(followers) != 2 {
		t.Errorf("Expected 2 followers, got %d", len(followers))
	}

	foundFirst := false
	foundSecond := false
	for _, f := range followers {
		if f == "192.168.1.10" {
			foundFirst = true
		}
		if f == "192.168.1.11" {
			foundSecond = true
		}
	}

	if !foundFirst || !foundSecond {
		t.Errorf("Missing expected followers in response: %v", followers)
	}
}

func TestPushTrack_ValidJSON(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":9999", "", provider, registry, trackQueue)

	jsonStr := []byte(`{"trackId":"test-track.mp3"}`)
	req := httptest.NewRequest(http.MethodPost, "/tracks", bytes.NewBuffer(jsonStr))
	rr := httptest.NewRecorder()

	server.PushTrack(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Verify track is in the queue
	if trackQueue.Len() != 1 {
		t.Errorf("Expected track queue to have 1 item, got %d", trackQueue.Len())
	}

	queuedTrack := trackQueue.PopFront()
	if queuedTrack != "test-track.mp3" {
		t.Errorf("Expected queued track to be 'test-track.mp3', got %s", queuedTrack)
	}
}

func TestPushTrack_InvalidJSON(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":9999", "", provider, registry, trackQueue)

	jsonStr := []byte(`{invalid-json}`) // Malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/tracks", bytes.NewBuffer(jsonStr))
	rr := httptest.NewRecorder()

	server.PushTrack(rr, req)

	// Since JSON parsing should fail, expect 400 Bad Request
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Verify the queue has NOT been modified
	if trackQueue.Len() != 0 {
		t.Errorf("Expected track queue to be empty, got %d", trackQueue.Len())
	}
}

func TestStartServer(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	// Assigning :0 lets the OS pick a random available port preventing address in use errors in tests
	server := NewWebServer(":0", "", provider, registry, trackQueue)

	go func() {
		server.StartServer()
	}()

	// Wait briefly to ensure it spins up without crashing
	time.Sleep(100 * time.Millisecond)
}

func TestBroadcastQueueState_SendsToConnections(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":0", "", provider, registry, trackQueue)

	// Simulate a connected SSE client
	ch := make(chan QueueState, 1)
	server.mu.Lock()
	server.connections = append(server.connections, ch)
	server.mu.Unlock()

	server.BroadcastQueueState("mock1", []string{"mock2"})

	select {
	case state := <-ch:
		if state.NowPlaying != "mock1" {
			t.Errorf("expected NowPlaying 'mock1', got '%s'", state.NowPlaying)
		}
		if len(state.QueueItems) != 1 || state.QueueItems[0] != "mock2" {
			t.Errorf("unexpected QueueItems: %v", state.QueueItems)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast")
	}
}

func TestBroadcastQueueState_SkipsFullChannels(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":0", "", provider, registry, trackQueue)

	// Simulate a slow client with a full channel
	ch := make(chan QueueState, 1)
	ch <- QueueState{NowPlaying: "stale"} // fill the buffer

	server.mu.Lock()
	server.connections = append(server.connections, ch)
	server.mu.Unlock()

	// Should not block even though the channel is full
	server.BroadcastQueueState("mock1", []string{})

	// Channel should still have the old message (non-blocking send skipped)
	state := <-ch
	if state.NowPlaying != "stale" {
		t.Errorf("expected stale message, got '%s'", state.NowPlaying)
	}
}

func TestBroadcastQueueState_EmptyQueueItems(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":0", "", provider, registry, trackQueue)

	ch := make(chan QueueState, 1)
	server.mu.Lock()
	server.connections = append(server.connections, ch)
	server.mu.Unlock()

	server.BroadcastQueueState("", nil)

	state := <-ch
	if state.NowPlaying != "" {
		t.Errorf("expected empty NowPlaying, got '%s'", state.NowPlaying)
	}
	if len(state.QueueItems) != 0 {
		t.Errorf("expected empty QueueItems, got %v", state.QueueItems)
	}
}

func TestStreamQueue_ConnectAndReceive(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":0", "", provider, registry, trackQueue)

	req := httptest.NewRequest(http.MethodGet, "/queue", nil)
	rr := httptest.NewRecorder()

	// Use a context that we can cancel to stop the streamer
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.StreamQueue(rr, req)
	}()

	// Wait for registration
	time.Sleep(50 * time.Millisecond)

	server.BroadcastQueueState("test.mp3", []string{"next.mp3"})

	// Give it a moment to process
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "test.mp3") {
		t.Errorf("expected response to contain test.mp3, got %s", responseBody)
	}
}

func TestStreamQueue_DisconnectRemovesChannel(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":0", "", provider, registry, trackQueue)

	req := httptest.NewRequest(http.MethodGet, "/queue", nil)
	rr := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.StreamQueue(rr, req)
	}()

	// Wait for registration
	time.Sleep(50 * time.Millisecond)

	server.mu.Lock()
	countBefore := len(server.connections)
	server.mu.Unlock()

	if countBefore != 1 {
		t.Fatalf("expected 1 connection, got %d", countBefore)
	}

	cancel() // Trigger disconnect
	wg.Wait()

	server.mu.Lock()
	countAfter := len(server.connections)
	server.mu.Unlock()

	if countAfter != 0 {
		t.Errorf("expected 0 connections after disconnect, got %d", countAfter)
	}
}
