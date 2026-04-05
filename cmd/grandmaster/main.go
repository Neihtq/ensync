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
	"ensync/internal/grandmaster/queue"
	"ensync/internal/grandmaster/sourceprovider"
	"ensync/internal/grandmaster/webservice"
)

const (
	logPrefix           = "[Main]"
	followerServicePort = ":8080"
	ntpPort             = ":9090"
	webPort             = ":9999"
)

func log(message string) {
	logging.Log(logPrefix, message)
}

func provideSourceProvider() sourceprovider.SourceProvider {
	return sourceprovider.NewAudioProvider()
}

func provideFollowersRegistry() *follower.FollowersRegistry {
	return follower.NewFollowersRegistry()
}

func provideTrackQueue() *queue.TrackQueue {
	return queue.NewTrackQueue()
}

func provideStreamer(
	followers *follower.FollowersRegistry,
	sourceProvider sourceprovider.SourceProvider,
	trackQueue *queue.TrackQueue,
) *audiostreamer.AudioStreamer {
	streamingInterval := 20 * time.Millisecond
	lookAhead := (2000 * time.Millisecond).Nanoseconds()
	sleepInterval := 100 * time.Millisecond
	return audiostreamer.NewAudioStreamer(followers, streamingInterval, lookAhead, sourceProvider, sleepInterval, trackQueue)
}

func provideClockSyncService() *clocksync.ClockSyncService {
	return clocksync.NewClockSyncService(ntpPort)
}

func provideDiscoveryService(registry *follower.FollowersRegistry) *discovery.DiscoveryService {
	return discovery.NewDiscoveryService(registry, ntpPort)
}

func provideWebserver(
	sourceProvider sourceprovider.SourceProvider,
	followersRegistry *follower.FollowersRegistry,
	trackQueue *queue.TrackQueue,
) *webservice.WebServer {
	return webservice.NewWebServer(webPort, sourceProvider, followersRegistry, trackQueue)
}

func main() {
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	followersRegistry := provideFollowersRegistry()

	log("Start NTP service")
	clockSyncService := provideClockSyncService()
	go clockSyncService.ExposeNTP(stop)

	log("Start AudioStreamLoop")
	sourceProvider := provideSourceProvider()
	trackQueue := provideTrackQueue()
	audioStreamer := provideStreamer(followersRegistry, sourceProvider, trackQueue)
	go audioStreamer.StreamAudioToAllLoop(stop)

	log("Start Discovery Service")
	discoveryService := provideDiscoveryService(followersRegistry)
	discoveryService.StartDiscovery()

	log("Start Web Server")
	webServer := provideWebserver(sourceProvider, followersRegistry, trackQueue)
	go webServer.StartServer()

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
