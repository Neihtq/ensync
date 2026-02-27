// Package clocksync server implements a server which the follower use to sync their time via NTP
package clocksync

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

const bufferSize = 8

func ExposeNTP(url string, stop chan struct{}) error {
	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer conn.Close()

	fmt.Println("Clock Sync Service listening on ", url)

	buffer := make([]byte, bufferSize)
	for {
		select {
		case <-stop:
			return nil
		default:
			numBytes, sender, err := conn.ReadFromUDP(buffer)
			receivedTime := time.Now().UnixNano()

			if err != nil {
				fmt.Println("Error reading: ", err)
				continue
			}
			timestamp := binary.BigEndian.Uint64(buffer[:numBytes])

			packet := make([]byte, 24)
			binary.BigEndian.PutUint64(packet[:8], uint64(timestamp))
			binary.BigEndian.PutUint64(packet[8:16], uint64(receivedTime))
			binary.BigEndian.PutUint64(packet[16:], uint64(time.Now().UnixNano())) // sendTime

			_, err = conn.WriteToUDP(packet, sender)
			if err != nil {
				fmt.Println("Error sending response:", err)
			}
		}
	}
}
