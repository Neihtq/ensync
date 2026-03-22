package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ensync/internal/follower/audio"
	"ensync/internal/follower/controlplane"
	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
	"ensync/internal/follower/visibility"

	"github.com/hashicorp/mdns"
)

const (
	audioPort     = ":9000"
	cpPort        = ":9001"
	cpPortInt     = 9001
	endpoint      = "/connections"
	lobbyEndpoint = "/visitors"
	lobbyPort     = ":9090"
)

func getLobbyAddress() string {
	data, err := os.ReadFile(".config")
	if err != nil {
		fmt.Println("error reading file:", err)
		return ""
	}

	address := strings.TrimSpace(string(data))
	fmt.Println("Read Lobby address:", address)

	return address
}

func connectLobby() {
	fmt.Println("Connecting to Lobby")
	serverAddr := getLobbyAddress() + lobbyPort + lobbyEndpoint
	fmt.Println("Full lobby address", serverAddr)
	for {
		err := visibility.JoinLobby(serverAddr, cpPort, endpoint)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

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

	// connectLobby()
	visibilityService := startVisibilityService()
	defer visibilityService.Shutdown()

	fmt.Println("Launch audio server.")
	audio.LaunchAudioServer(audioPort, ipProvider, mirrorClock, stop)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)

	fmt.Println("Exit.")
}
