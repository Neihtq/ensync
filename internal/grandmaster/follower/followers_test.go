package follower

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestAddFollowerAddsFollowerWithURLs(t *testing.T) {
	followers := NewFollowers()

	address := "http://127.0.0.1"
	heartbeatPort := "5000"
	heartbeatURL := address + ":" + heartbeatPort
	audioPort := "5001"
	audioURL := address + ":" + audioPort

	jsonBody := []byte(`{"address": "` + address + `", "heartbeatPort": "` + heartbeatPort + `", "audioPort": "` + audioPort + `"}`)
	request := httptest.NewRequest("POST", "/followers", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	followers.AddFollower(writer, request)

	if len(followers.Followers) != 1 {
		t.Error("Expected one HeartbeatUrl in Followers. URLs: %u", followers.Followers)
	}
	if followers.Followers[address].HeartbeatURL != heartbeatURL {
		t.Error("Expected HeartbeatUrl to be "+heartbeatURL+". Got %s instead", heartbeatURL, followers.Followers[address].HeartbeatURL)
	}
	if followers.Followers[address].AudioURL != audioURL {
		t.Error("Expected AudioUrl to be "+audioURL+". Got %s instead", audioURL, followers.Followers[address].AudioURL)
	}
}
