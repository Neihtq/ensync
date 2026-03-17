package visibility

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/mdns"
)

func TestExposeMDNS(t *testing.T) {
	t.Skip("Skipping mDNS test")
	// arange
	port := 9999
	info := []string{"/connections"}

	// act
	server, _ := ExposeMDNS(port, info)
	defer server.Shutdown()
	time.Sleep(10 * time.Millisecond)

	// assert
	entriesCh := make(chan *mdns.ServiceEntry, 1)

	go func() {
		params := mdns.DefaultParams("_ensync._tcp")
		params.Entries = entriesCh
		params.Timeout = 100 * time.Millisecond
		mdns.Query(params)
		close(entriesCh)
	}()

	found := false
	for entry := range entriesCh {
		if entry.Port == port {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("ExposeMDNS failed: mDNS service not discovered.")
	}
}

func TestJoinLobby(t *testing.T) {
	// arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		var data map[string]string
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	addr := server.URL[len("http://"):]
	endpoint := "/connections"

	// act
	err := JoinLobby(addr, "8080", endpoint)
	// assert
	if err != nil {
		t.Errorf("JoinLobby failed: %v", err)
	}
}
