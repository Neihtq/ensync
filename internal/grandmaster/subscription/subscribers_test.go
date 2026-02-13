package subscription

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestSubscribeAddUrlToSubscribers(t *testing.T) {
	subscribers := &Subscribers{}
	url := "http://127.0.0.1:5000"
	jsonBody := []byte(`{"url": "` + url + `"}`)
	request := httptest.NewRequest("POST", "/subscribe", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	subscribers.Subscribe(writer, request)

	if len(subscribers.Urls) != 1 {
		t.Error("Expected one URL in Subscribers. URLs: %u", subscribers.Urls)
	}

	if subscribers.Urls[0] != url {
		t.Error("Expected URL to be "+url+". Got %s instead", url, subscribers.Urls[0])
	}
}
