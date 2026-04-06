package navidrome

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNavidromeClient_Ping_Success(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify some common params
		query := r.URL.Query()
		if query.Get("f") != "json" {
			t.Errorf("expected format json, got %s", query.Get("f"))
		}
		if query.Get("u") == "" {
			t.Errorf("expected username, got empty")
		}

		resp := ResponseWrapper{
			SubsonicResponse: SubsonicResponse{
				Status:  "ok",
				Version: "1.16.1",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Set env vars for the client
	os.Setenv("NAVIDROME_URL", server.URL)
	os.Setenv("NAVIDROME_USER", "testuser")
	os.Setenv("NAVIDROME_PASSWORD", "testpass")
	defer os.Unsetenv("NAVIDROME_URL")
	defer os.Unsetenv("NAVIDROME_USER")
	defer os.Unsetenv("NAVIDROME_PASSWORD")

	client := NewNavidromeClient()
	err := client.Ping()

	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestNavidromeClient_Ping_Failure(t *testing.T) {
	// Setup mock server returning an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ResponseWrapper{
			SubsonicResponse: SubsonicResponse{
				Status: "failed",
				Error: &APIError{
					Code:    40,
					Message: "Wrong username or password",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	os.Setenv("NAVIDROME_URL", server.URL)
	os.Setenv("NAVIDROME_USER", "baduser")
	os.Setenv("NAVIDROME_PASSWORD", "badpass")
	defer os.Unsetenv("NAVIDROME_URL")
	
	client := NewNavidromeClient()
	err := client.Ping()

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	expectedError := "API error 40: Wrong username or password"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestNavidromeClient_HTTPError(t *testing.T) {
	client := NewNavidromeClient()
	client.BaseURL = "http://non-existent-url-12345.com"
	
	err := client.Ping()
	if err == nil {
		t.Fatal("expected HTTP error but got nil")
	}
}
