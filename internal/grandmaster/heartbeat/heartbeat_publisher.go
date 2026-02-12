// Package heartbeat for sending time stamps
package heartbeat

import (
	"fmt"
	"time"

	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/subscription"
)

const (
	logPrefix = "[HeartbeatPublisher]"
	interval  = 5 * time.Second
)

type HeartbeatPublisher struct {
	Subs *subscription.Subscribers
}

func (publisher *HeartbeatPublisher) SendHeartbeatToAll() {
	for _, url := range publisher.Subs.Urls {
		fmt.Println("Send hearbeat to " + url)
		logging.Log(logPrefix, "Send hearbeat to "+url)
	}
}

func (publisher *HeartbeatPublisher) HeartbeatLoop() {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			publisher.SendHeartbeatToAll()
		}
	}
}
