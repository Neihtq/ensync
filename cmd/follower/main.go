package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ensync/internal/follower/middleware"
)

const udpPort = ":9000"

func main() {
	stop := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ipProvider := middleware.RealIPProvider{}
	endpointProvider := middleware.SubscriptionEndpointProvider{}

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		close(stop)
	}()

	fmt.Println("Starting Application.")
	middleware.SubscribeAndExpose(udpPort, stop, ipProvider, endpointProvider)
}
