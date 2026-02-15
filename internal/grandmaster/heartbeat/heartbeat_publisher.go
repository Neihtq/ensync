// Package heartbeat for sending time stamps
package heartbeat

import (
	"fmt"
	"log"
	"net"
	"time"

	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/subscription"
)

const (
	logPrefix = "[HeartbeatPublisher]"
	interval  = 100 * time.Millisecond
)

type HeartbeatPublisher struct {
	Subs *subscription.Subscribers
}

func SendHeartbeat(url string) {
	fmt.Println("Send hearbeat to " + url)
	logging.Log(logPrefix, "Send hearbeat to "+url)

	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
	}
	logging.Log(logPrefix, "Address: "+addr.String())

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	message := []byte("heartbeat")
	_, err = conn.Write(message)
	if err != nil {
		logging.Log(logPrefix, "Error sending message: "+err.Error())
		return
	}

	logging.Log(logPrefix, "Send message to "+url)
}

func (publisher *HeartbeatPublisher) SendHeartbeatToAll() {
	for _, url := range publisher.Subs.Urls {
		go SendHeartbeat(url)
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
