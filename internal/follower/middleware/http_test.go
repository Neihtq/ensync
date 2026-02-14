package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func StartMockHTTPServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Errorf("Expected POST request, got %s", request.Method)
		}

		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json header, got %s", request.Header.Get("Content-Type"))
		}

		writer.WriteHeader(http.StatusNoContent)
	}))
}

func TestPost(t *testing.T) {
	server := StartMockHTTPServer(t)
	defer server.Close()

	testData := map[string]string{"url": "TestUrl1:8080"}
	err := post(testData, server.URL)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}
