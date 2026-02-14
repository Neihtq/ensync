package middleware

import (
	"fmt"
	"net"
	"os"
)

func expose(port string, stop chan struct{}) error {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting listener:", err)
		return err
	}
	defer conn.Close()

	fmt.Printf("UDP server listening on %s", port)

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
			numBytes, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading:", err)
				continue
			}

			fmt.Printf("Received %d bytes from %s: %s\n", numBytes, remoteAddr, string(buffer[:numBytes]))
		}
	}
}

func SubscribeAndExpose(
	port string,
	stop chan struct{},
	ipProvider IPProvider,
	endpointProvider EndpointProvider,
) {
	localIP := ipProvider.GetIP().String()
	fmt.Printf("IP Address: %s", localIP)

	grandmasterEndpoint := endpointProvider.GetEndpoint()
	localUDPAddr := "http://" + localIP + port
	data := map[string]string{"url": localUDPAddr}
	err := post(data, grandmasterEndpoint)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	expose(port, stop)
}
