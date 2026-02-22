// Package heartbeat for sending time stamps
package heartbeat

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/subscription"
)

const (
	logPrefix    = "[HeartbeatPublisher]"
	interval     = 100 * time.Millisecond
	envelopeSize = 8
)

type HeartbeatPublisher struct {
	Subs *subscription.Subscribers
}

func SendHeartbeat(url string) {
	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	envelope := make([]byte, envelopeSize)
	timestamp := time.Now().UnixNano()
	binary.BigEndian.PutUint64(envelope, uint64(timestamp))
	_, err = conn.Write(envelope)
	if err != nil {
		logging.Log(logPrefix, "Error sending message: "+err.Error())
		return
	}
}

func (publisher *HeartbeatPublisher) SendHeartbeatToAll() {
	publisher.Subs.RLock()
	defer publisher.Subs.RUnlock()
	for _, url := range publisher.Subs.HeartbeatURLs {
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
