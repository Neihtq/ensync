package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/follower/audio"
	"ensync/internal/follower/clocksync"
	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

const (
	audioPort     = "9000"
	heartbeatPort = "9001"
	ntpPort       = ":9090"
)

func main() {
	stop := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	mirrorClock := mirrorclock.NewMirrorClock()
	ipProvider := middleware.RealIPProvider{}

	ntpAddressProvider := middleware.NTPAddressProvider{}
	serverURL := "127.0.0.1"
	clockSync := clocksync.NewClockSync(mirrorClock, ntpAddressProvider.BuildAddress(serverURL, ntpPort))
	interval := 10 * time.Millisecond
	go clockSync.RunClockSync(interval, stop)

	fmt.Println("Launch audio server.")
	audio.LaunchAudioServer(audioPort, ipProvider, mirrorClock, stop)

	<-sigChan
	fmt.Println("\nShutting down...")
	close(stop)

	fmt.Println("Exit.")
}
