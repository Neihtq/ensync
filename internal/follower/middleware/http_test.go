package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func StartMockHTTPServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
}

func TestPost(t *testing.T) {
	// arrange
	server := StartMockHTTPServer(t)
	defer server.Close()

	testData := map[string]string{"url": "TestUrl1:8080"}

	// act
	err := Post(testData, server.URL)
	// assert
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

func TestDelete(t *testing.T) {
	// arrange
	server := StartMockHTTPServer(t)

	// act
	err := Delete(server.URL, "testParam")
	// assert
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}
