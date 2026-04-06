package visibility

import (
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
