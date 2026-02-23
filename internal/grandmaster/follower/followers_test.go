package follower

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddFollowerAddsFollowerWithURLs(t *testing.T) {
	// arrange
	followers := NewFollowers()

	address := "127.0.0.1"
	heartbeatPort := "5000"
	heartbeatURL := address + ":" + heartbeatPort
	audioPort := "5001"
	audioURL := address + ":" + audioPort

	jsonBody := []byte(`{"address": "` + address + `", "heartbeatPort": "` + heartbeatPort + `", "audioPort": "` + audioPort + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/followers", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	// act
	followers.AddFollower(writer, request)

	// assert
	if writer.Code != http.StatusCreated && writer.Code != http.StatusOK {
		t.Errorf("Expected status 201/200, got %d", writer.Code)
	}
	if len(followers.Followers) != 1 {
		t.Error("Expected one Followers. But got: %u", followers.Followers)
	}
	if followers.Followers[address].HeartbeatURL != heartbeatURL {
		t.Error("Expected HeartbeatUrl to be "+heartbeatURL+". Got %s instead", heartbeatURL, followers.Followers[address].HeartbeatURL)
	}
	if followers.Followers[address].AudioURL != audioURL {
		t.Error("Expected AudioUrl to be "+audioURL+". Got %s instead", audioURL, followers.Followers[address].AudioURL)
	}
}

func TestRemoveFollowerRemoveFollower(t *testing.T) {
	// arrange
	followers := NewFollowers()

	address := "127.0.0.1"
	heartbeatPort := "5000"
	heartbeatURL := address + ":" + heartbeatPort
	audioPort := "5001"
	audioURL := address + ":" + audioPort
	followers.Followers[address] = NewFollower(heartbeatURL, audioURL)

	request := httptest.NewRequest(http.MethodDelete, "/followers/"+address, nil)
	request.SetPathValue("address", address)
	writer := httptest.NewRecorder()

	// act
	followers.RemoveFollower(writer, request)

	// assert
	if writer.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", writer.Code)
	}
	if len(followers.Followers) > 0 {
		t.Errorf("Expected no Followers. But got: %v", followers.Followers)
	}
}
