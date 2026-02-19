package subscription

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestSubscribeAddUrlsToSubscribers(t *testing.T) {
	subscribers := &Subscribers{}

	address := "http://127.0.0.1"
	heartbeatPort := "5000"
	heartbeatURL := address + ":" + heartbeatPort
	audioPort := "5001"
	audioURL := address + ":" + audioPort

	jsonBody := []byte(`{"address": "` + address + `", "heartbeatPort": "` + heartbeatPort + `", "audioPort": "` + audioPort + `"}`)
	request := httptest.NewRequest("POST", "/subscribe", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	subscribers.Subscribe(writer, request)

	if len(subscribers.HeartbeatURLs) != 1 {
		t.Error("Expected one HeartbeatUrl in Subscribers. URLs: %u", subscribers.HeartbeatURLs)
	}
	if subscribers.HeartbeatURLs[0] != heartbeatURL {
		t.Error("Expected HeartbeatUrl to be "+heartbeatURL+". Got %s instead", heartbeatURL, subscribers.HeartbeatURLs[0])
	}

	if len(subscribers.AudioURLs) != 1 {
		t.Error("Expected one AudioUrl in Subscribers. URLs: %u", subscribers.AudioURLs)
	}
	if subscribers.AudioURLs[0] != audioURL {
		t.Error("Expected AudioUrl to be "+audioURL+". Got %s instead", audioURL, subscribers.AudioURLs[0])
	}
}
