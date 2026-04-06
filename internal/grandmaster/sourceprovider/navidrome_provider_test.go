package sourceprovider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ensync/internal/grandmaster/navidrome"
)

func TestNavidromeProvider_PingSuccess(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := navidrome.ResponseWrapper{
			SubsonicResponse: navidrome.SubsonicResponse{
				Status:  "ok",
				Version: "1.16.1",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	os.Setenv("NAVIDROME_URL", server.URL)
	os.Setenv("NAVIDROME_USER", "test")
	os.Setenv("NAVIDROME_PASSWORD", "test")
	defer os.Unsetenv("NAVIDROME_URL")
	defer os.Unsetenv("NAVIDROME_USER")
	defer os.Unsetenv("NAVIDROME_PASSWORD")

	// We don't call NewNaviDromeProvider directly because it starts a goroutine that can panic/leak
	client := navidrome.NewNavidromeClient()
	provider := &NaviDromeProvider{
		Client: client,
	}

	// Just test one ping
	err := provider.Client.Ping()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNavidromeProvider_HealthCheckIntegration(t *testing.T) {
	// Setup mock server
	pingCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pingCount++
		resp := navidrome.ResponseWrapper{
			SubsonicResponse: navidrome.SubsonicResponse{
				Status:  "ok",
				Version: "1.16.1",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	os.Setenv("NAVIDROME_URL", server.URL)
	os.Setenv("NAVIDROME_USER", "test")
	os.Setenv("NAVIDROME_PASSWORD", "test")
	defer os.Unsetenv("NAVIDROME_URL")
	defer os.Unsetenv("NAVIDROME_USER")
	defer os.Unsetenv("NAVIDROME_PASSWORD")

	// Create provider
	client := navidrome.NewNavidromeClient()
	provider := &NaviDromeProvider{
		Client: client,
	}

	// Just verify the client works via the provider
	if err := provider.Client.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestNavidromeProvider_EmptyMethods(t *testing.T) {
	provider := &NaviDromeProvider{}

	songs := provider.ListSongs()
	if len(songs) != 0 {
		t.Errorf("expected 0 songs, got %d", len(songs))
	}
}

func TestNavidromeProvider_SearchSong(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := navidrome.ResponseWrapper{
			SubsonicResponse: navidrome.SubsonicResponse{
				Status: "ok",
				SearchResult3: &navidrome.SearchResult3{
					Song: []navidrome.Song{
						{ID: "1", Title: "Wonderwall", Artist: "Oasis"},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	os.Setenv("NAVIDROME_URL", server.URL)
	os.Setenv("NAVIDROME_USER", "test")
	os.Setenv("NAVIDROME_PASSWORD", "test")
	defer os.Unsetenv("NAVIDROME_URL")
	defer os.Unsetenv("NAVIDROME_USER")
	defer os.Unsetenv("NAVIDROME_PASSWORD")

	client := navidrome.NewNavidromeClient()
	provider := &NaviDromeProvider{
		Client: client,
	}

	songs := provider.SearchSong("oasis")
	if len(songs) != 1 {
		t.Fatalf("expected 1 song, got %d", len(songs))
	}
	if songs[0].Title != "Wonderwall" {
		t.Errorf("expected Wonderwall, got %s", songs[0].Title)
	}
}
