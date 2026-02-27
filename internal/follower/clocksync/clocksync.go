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
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	clockSync.mu.Lock()
	clockSync.ListeningAddr = conn.LocalAddr().String()
	clockSync.mu.Unlock()

	fmt.Println("Listening to server on port ", port)
	buffer := make([]byte, 24)
	for {
		select {
		case <-stop:
			return
		default:
			numBytes, _, err := conn.ReadFromUDP(buffer)
			receivedTime := uint64(time.Now().UnixNano())
			if err != nil {
				fmt.Println("Error reading:", err)
				continue
			}
			timeStamps := []uint64{}
			for start := 0; start+timeStampSize <= numBytes; start += timeStampSize {
				ts := binary.BigEndian.Uint64(buffer[start : start+timeStampSize])
				timeStamps = append(timeStamps, ts)
			}
			timeStamps = append(timeStamps, receivedTime)
			clockSync.Clock.SyncTime(timeStamps)
		}
	}
}
