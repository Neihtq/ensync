package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/follower/audio"
	"ensync/internal/follower/middleware"
)

const (
	audioPort = "9000"
	udpPort   = "9001"
)

func main() {
	stop := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ipProvider := middleware.RealIPProvider{}
	endpointProvider := middleware.SubscriptionEndpointProvider{}

	fmt.Println("Starting Application.")
	go middleware.SubscribeAndExpose(udpPort, audioPort, stop, ipProvider, endpointProvider)

	fmt.Println("Launch audio server.")
	audio.LaunchAudioServer(audioPort, ipProvider, stop)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)

	time.Sleep(time.Millisecond * 500)
	fmt.Println("Exit.")
}
