package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/grandmaster/audiostreamer"
	"ensync/internal/grandmaster/clocksync"
	"ensync/internal/grandmaster/discovery"
	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/sourceprovider"
)

const (
	logPrefix           = "[Main]"
	followerServicePort = ":8080"
	ntpPort             = ":9090"
)

func log(message string) {
	logging.Log(logPrefix, message)
}

func provideSourceProvider() sourceprovider.SourceProvider {
	return sourceprovider.NewAudioProvider()
}

func provideFollowers() *follower.Followers {
	return follower.NewFollowers()
}

func provideStreamer(followers *follower.Followers, sourceProvider sourceprovider.SourceProvider) *audiostreamer.AudioStreamer {
	log("Initialize AudioStreamer")
	streamingInterval := 20 * time.Millisecond
	lookAhead := (2000 * time.Millisecond).Nanoseconds()
	sleepInterval := 100 * time.Millisecond
	return audiostreamer.NewAudioStreamer(followers, streamingInterval, lookAhead, sourceProvider, sleepInterval)
}

func startDiscoveryService(followers *follower.Followers) {
	log("Start Discovery Service")
	discoveryService := discovery.NewDiscoveryService(followers, ntpPort)
	discoveryService.Discover()
}

func main() {
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	followers := provideFollowers()
	sourceProvider := provideSourceProvider()
	audioStreamer := provideStreamer(followers, sourceProvider)

	log("Start NTP service")
	go clocksync.ExposeNTP(ntpPort, stop)

	log("Start AudioStreamLoop with sending interval")
	go audioStreamer.StreamAudioToAllLoop(stop)

	startDiscoveryService(followers)

	var input string
	fmt.Println("Continue? [y]es")
	fmt.Scan(&input)
	if input == "y" {
		filePath := "./assets/test.mp3"
		audioStreamer.AddToQueue(filePath)
	}

	<-sigChan
	log("Shutting down...")
	time.Sleep(time.Millisecond * 500)
	log("Exit.")
}
