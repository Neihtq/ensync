package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/follower/audio"
	"ensync/internal/follower/controlplane"
	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
	"ensync/internal/follower/visibility"
)

const (
	audioPort = ":9000"
	cpPort    = ":9001"
	cpPortInt = 9001
	ntpPort   = ":9090"
	endpoint  = "/connections"
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

	fmt.Println("Connecting to Lobby")
	serverAddr := "http://127.0.0.1/visitors"
	for {
		err := visibility.JoinLobby(serverAddr, cpPort, endpoint)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	for cp.ClockSync == nil {
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Launch audio server.")
	audio.LaunchAudioServer(audioPort, ipProvider, mirrorClock, stop)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)

	fmt.Println("Exit.")
}
