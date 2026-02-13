package heartbeat

import (
	"testing"

	"ensync/internal/grandmaster/subscription"
)

func TestHeartbeatPublisherProcessesEachUrl(t *testing.T) {
	urls := []string{"testUrl1", "testUrl2", "testUrl3"}
	subscribers := subscription.Subscribers{Urls: urls}

	heartbeatPublisher := &HeartbeatPublisher{Subs: &subscribers}
	heartbeatPublisher.SendHeartbeatToAll()
}
