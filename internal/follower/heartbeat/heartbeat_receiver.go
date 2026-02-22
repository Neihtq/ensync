// Package heartbeat implements receiver logic to get heartbeats from grandmaster
package heartbeat

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

type HeartbeatReceiver struct {
	IPAddress   string
	Port        string
	URL         string
	MirrorClock *mirrorclock.MirrorClock
}

func NewHeartbeatReceiver(
	port string,
	ipProvider middleware.IPProvider,
	mirrorClock *mirrorclock.MirrorClock,
) *HeartbeatReceiver {
	localIPAddress := ipProvider.GetIP().String()
	url := localIPAddress + ":" + port

	return &HeartbeatReceiver{
		IPAddress:   localIPAddress,
		Port:        port,
		URL:         url,
		MirrorClock: mirrorClock,
	}
}

func (receiver *HeartbeatReceiver) expose(stop chan struct{}) error {
	addr, err := net.ResolveUDPAddr("udp", receiver.URL)
	if err != nil {
		fmt.Println("Error resolving address for heartbeat server:\n", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting listener for heartbeat server:\n", err)
		return err
	}
	defer conn.Close()

	fmt.Printf("Heartbeat server listening on %s\n", receiver.URL)

	go func() {
		<-stop
		conn.Close()
	}()

	buffer := make([]byte, 1024)
	for {
		select {
		case <-stop:
			return nil
		default:
			numBytes, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				select {
				case <-stop:
					return nil
				default:
					fmt.Println("Error reading:", err)
					continue
				}
			}
			timestamp := binary.BigEndian.Uint64(buffer[:numBytes])
			receiver.MirrorClock.UpdateOffset(timestamp)
		}
	}
}

func (receiver *HeartbeatReceiver) SubscribeAndExpose(
	audioPort string,
	stop chan struct{},
	endpointProvider middleware.EndpointProvider,
) {
	grandmasterEndpoint := endpointProvider.GetEndpoint()
	data := map[string]string{
		"address":       receiver.IPAddress,
		"heartbeatPort": receiver.Port,
		"audioPort":     audioPort,
	}
	err := middleware.Post(data, grandmasterEndpoint)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	receiver.expose(stop)
}
