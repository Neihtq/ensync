// Package clocksync follower implements follower logic to perform NTP exchange with the server
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
	mu        sync.Mutex
	Clock     *mirrorclock.MirrorClock
	Conn      *net.UDPConn
	Interval  time.Duration
	Heartbeat *net.TCPConn
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

	tcpAddr, err := net.ResolveTCPAddr("tcp", serverURL)
	if err != nil {
		log.Fatal(err)
	}
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	return &ClockSync{
		Clock:     clock,
		Conn:      conn,
		Interval:  100 * time.Millisecond,
		Heartbeat: tcpConn,
	}
}

func (clockSync *ClockSync) SendNTPRequest() error {
	followerSendTime := time.Now().UnixNano()
	packet := make([]byte, timeStampSize)
	binary.BigEndian.PutUint64(packet, uint64(followerSendTime))
	_, err := clockSync.Conn.Write(packet)
	if err != nil {
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
				fmt.Println("Host not reachable. Closing connection.", err)
				return
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

func (clockSync *ClockSync) RunClockSync(stop chan struct{}) {
	go clockSync.ReceiveNTPPackets(stop)
	ticker := time.NewTicker(clockSync.Interval)
	defer ticker.Stop()
	defer clockSync.Heartbeat.Close()
	defer clockSync.Conn.Close()

	buf := make([]byte, 1)
	for {
		select {
		case <-stop:
			fmt.Println("Stopping Clocksync...")
			return
		case <-ticker.C:
			_, dead := clockSync.Heartbeat.Write(buf)
			if dead != nil {
				fmt.Println("Hearbeat service unavailable:", dead)
				return
			}
			err := clockSync.SendNTPRequest()
			if err != nil {
				fmt.Println("ClockSync service unavailable:", err)
				return
			}
		}
	}
}
