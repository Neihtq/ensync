package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ensync/internal/follower/audio"
	"ensync/internal/follower/heartbeat"
	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

const (
	audioPort     = "9000"
	heartbeatPort = "9001"
)

func main() {
	stop := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	mirrorClock := mirrorclock.NewMirrorClock()
	ipProvider := middleware.RealIPProvider{}
	endpointProvider := middleware.FollowersEndpointProvider{}
	heartbeatReceiver := heartbeat.NewHeartbeatReceiver(heartbeatPort, ipProvider, mirrorClock)

	fmt.Println("Starting Application.")
	go heartbeatReceiver.SubscribeAndExpose(audioPort, stop, endpointProvider)

	fmt.Println("Launch audio server.")
	audio.LaunchAudioServer(audioPort, ipProvider, stop)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)

	err := middleware.Delete(endpointProvider.GetEndpoint(), ipProvider.GetIP().String())
	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Println("Exit.")
}
