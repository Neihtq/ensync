package middleware

import (
	"fmt"
	"net"
	"os"
)

func expose(url string, stop chan struct{}) error {
	addr, err := net.ResolveUDPAddr("udp", url)
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

	fmt.Printf("Heartbeat server listening on %s\n", url)

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
			_, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				select {
				case <-stop:
					return nil
				default:
					fmt.Println("Error reading:", err)
					continue
				}
			}
		}
	}
}

func SubscribeAndExpose(
	heartbeatPort string,
	audioPort string,
	stop chan struct{},
	ipProvider IPProvider,
	endpointProvider EndpointProvider,
) {
	localIPAddress := ipProvider.GetIP().String()
	fmt.Printf("IP Address: %s", localIPAddress)

	grandmasterEndpoint := endpointProvider.GetEndpoint()
	data := map[string]string{"address": localIPAddress, "heartbeatPort": heartbeatPort, "audioPort": audioPort}
	err := post(data, grandmasterEndpoint)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	expose(localIPAddress+":"+heartbeatPort, stop)
}
