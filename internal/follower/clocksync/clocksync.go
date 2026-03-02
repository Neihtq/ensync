// Package clocksync implements logic for NTP
package clocksync

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"ensync/internal/follower/mirrorclock"
)

const (
	timeStampSize = 8
	port          = ":9999"
)

type ClockSync struct {
	mu            sync.Mutex
	Clock         *mirrorclock.MirrorClock
	Conn          *net.UDPConn
	ListeningAddr string
}

func NewClockSync(clock *mirrorclock.MirrorClock, serverURL string) *ClockSync {
	addr, err := net.ResolveUDPAddr("udp", serverURL)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	return &ClockSync{
		Clock: clock,
		Conn:  conn,
	}
}

func (clockSync *ClockSync) SendNTPRequest() error {
	t1 := time.Now().UnixNano()
	packet := make([]byte, timeStampSize)
	binary.BigEndian.PutUint64(packet, uint64(t1))
	_, err := clockSync.Conn.Write(packet)
	if err != nil {
		fmt.Println("Error sending NTP request", err)
		return err
	}

	return nil
}

func (clockSync *ClockSync) ReceiveNTPPackets(stop chan struct{}) {
	buffer := make([]byte, 24)
	for {
		select {
		case <-stop:
			return
		default:
			numBytes, _, err := clockSync.Conn.ReadFromUDP(buffer)
			receivedTime := time.Now().UnixNano()
			if err != nil {
				fmt.Println("Error reading:", err)
				continue
			}
			timeStamps := []int64{}
			for start := 0; start+timeStampSize <= numBytes; start += timeStampSize {
				ts := int64(binary.BigEndian.Uint64(buffer[start : start+timeStampSize]))
				timeStamps = append(timeStamps, ts)
			}
			timeStamps = append(timeStamps, receivedTime)
			clockSync.Clock.SyncTime(timeStamps)
		}
	}
}

func (clockSync *ClockSync) RunClockSync(interval time.Duration, stop chan struct{}) {
	go clockSync.ReceiveNTPPackets(stop)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			fmt.Println("Stopping Clocksync...")
			return
		case <-ticker.C:
			clockSync.SendNTPRequest()
		}
	}
}
