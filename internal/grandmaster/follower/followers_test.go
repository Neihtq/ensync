package follower

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestCreateFollowerAddUrlsToFollowers(t *testing.T) {
	followers := &Followers{}

	address := "http://127.0.0.1"
	heartbeatPort := "5000"
	heartbeatURL := address + ":" + heartbeatPort
	audioPort := "5001"
	audioURL := address + ":" + audioPort

	jsonBody := []byte(`{"address": "` + address + `", "heartbeatPort": "` + heartbeatPort + `", "audioPort": "` + audioPort + `"}`)
	request := httptest.NewRequest("POST", "/followers", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	followers.CreateFollower(writer, request)

	if len(followers.HeartbeatURLs) != 1 {
		t.Error("Expected one HeartbeatUrl in Followers. URLs: %u", followers.HeartbeatURLs)
	}
	if followers.HeartbeatURLs[0] != heartbeatURL {
		t.Error("Expected HeartbeatUrl to be "+heartbeatURL+". Got %s instead", heartbeatURL, followers.HeartbeatURLs[0])
	}

	if len(followers.AudioURLs) != 1 {
		t.Error("Expected one AudioUrl in Followers. URLs: %u", followers.AudioURLs)
	}
	if followers.AudioURLs[0] != audioURL {
		t.Error("Expected AudioUrl to be "+audioURL+". Got %s instead", audioURL, followers.AudioURLs[0])
	}
}
