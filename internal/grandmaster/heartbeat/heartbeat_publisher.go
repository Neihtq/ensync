// Package heartbeat for sending time stamps
package heartbeat

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/logging"
)

const (
	logPrefix     = "[HeartbeatPublisher]"
	constInterval = 100 * time.Millisecond
	envelopeSize  = 8
)

type HeartbeatPublisher struct {
	Followers *follower.Followers
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
	publisher.Followers.RLock()
	defer publisher.Followers.RUnlock()
	for _, f := range publisher.Followers.Followers {
		url := f.HeartbeatURL
		go SendHeartbeat(url)
	}
}

func (publisher *HeartbeatPublisher) HeartbeatLoop(interval time.Duration, stop chan struct{}) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-stop:
			logging.Log(logPrefix, "Stopping heartbeat loop...")
			return
		case <-ticker.C:
			publisher.SendHeartbeatToAll()
		}
	}
}
