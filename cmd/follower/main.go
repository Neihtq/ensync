package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ensync/internal/follower/audio"
	"ensync/internal/follower/controlplane"
	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

const (
	audioPort = ":9000"
	cpPort    = ":9001"
	ntpPort   = ":9090"
)

func main() {
	stop := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	mirrorClock := mirrorclock.NewMirrorClock()
	ipProvider := middleware.RealIPProvider{}

	fmt.Println("Start ControlPlane")
	cp := controlplane.NewControlPlaneService(mirrorClock, audioPort, stop)
	go cp.StartService(cpPort)

	fmt.Println("Launch audio server.")
	audio.LaunchAudioServer(audioPort, ipProvider, mirrorClock, stop)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)

	fmt.Println("Exit.")
}
