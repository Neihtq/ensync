// Package clocksync server implements a server which the follower use to sync their time via NTP
package clocksync

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

const timeStampSize = 8

func ExposeNTP(port string, stop chan struct{}) error {
	addr, err := net.ResolveUDPAddr("udp", port)
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

	fmt.Println("Clock Sync Service listening on ", conn.LocalAddr().String())

	buffer := make([]byte, 24)
	for {
		select {
		case <-stop:
			return nil
		default:
			numBytes, sender, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading: ", err)
				continue
			}
			receivedTime := time.Now().UnixNano()

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
