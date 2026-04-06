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
	"ensync/internal/follower/visibility"

	"github.com/hashicorp/mdns"
)

const (
	audioPort     = ":65532"
	cpPort        = ":65531"
	cpPortInt     = 65531
	endpoint      = "/connections"
	lobbyEndpoint = "/visitors"
)

func startVisibilityService() *mdns.Server {
	fmt.Println("Start Visibility Service")
	info := []string{"/connections"}
	visibilityService, err := visibility.ExposeMDNS(cpPortInt, info)
	if err != nil {
		fmt.Println("Exposing mDNS failed", err.Error())
	}

	return visibilityService
}

func main() {
	stop := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	mirrorClock := mirrorclock.NewMirrorClock()
	ipProvider := middleware.RealIPProvider{}

	cp := controlplane.NewControlPlaneService(mirrorClock, audioPort, stop)
	go cp.StartService(cpPort)

	visibilityService := startVisibilityService()
	defer visibilityService.Shutdown()

	fmt.Println("Launch audio server.")
	audio.LaunchAudioServer(audioPort, ipProvider, mirrorClock, stop)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)

	fmt.Println("Exit.")
}
