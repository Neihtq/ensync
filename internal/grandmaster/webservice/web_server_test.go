package webservice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/queue"
	"ensync/internal/grandmaster/sourceprovider"
)

func TestNewWebServer(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	port := ":9999"

	server := NewWebServer(port, provider, registry, trackQueue)
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

func TestListSongs(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()
	server := NewWebServer(":9999", provider, registry, trackQueue)

	req := httptest.NewRequest(http.MethodGet, "/songs", nil)
	rr := httptest.NewRecorder()

	server.listSongs(rr, req)

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

	if len(songs) != 2 || songs[0] != "track1" || songs[1] != "track2" {
		t.Errorf("Unexpected songs list: %v", songs)
	}
}

func TestListFollowers(t *testing.T) {
	provider := &sourceprovider.MockSourceProvider{}
	registry := follower.NewFollowersRegistry()
	trackQueue := queue.NewTrackQueue()

	registry.RegisterFollower("192.168.1.10", "8000")
	registry.RegisterFollower("192.168.1.11", "8000")

	server := NewWebServer(":9999", provider, registry, trackQueue)

	req := httptest.NewRequest(http.MethodGet, "/followers", nil)
	rr := httptest.NewRecorder()

	server.listFollowers(rr, req)

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
	server := NewWebServer(":9999", provider, registry, trackQueue)

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
	server := NewWebServer(":9999", provider, registry, trackQueue)

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
	server := NewWebServer(":0", provider, registry, trackQueue)

	go func() {
		server.StartServer()
	}()

	// Wait briefly to ensure it spins up without crashing
	time.Sleep(100 * time.Millisecond)
}
