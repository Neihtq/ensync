package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/grandmaster/audiostreamer"
	"ensync/internal/grandmaster/clocksync"
	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/logging"
)

const (
	logPrefix           = "[Main]"
	followerServicePort = ":8080"
	ntpPort             = ":9090"
)

func log(message string) {
	logging.Log(logPrefix, message)
}

func initializeFixtures() (*follower.Followers, *audiostreamer.AudioStreamer) {
	log("Initialize registry of Subscribers")
	followers := follower.NewFollowers()

	log("Initialize AudioStreamer")
	interval := 20 * time.Millisecond
	lookAhead := (2000 * time.Millisecond).Nanoseconds()
	audioProvider := &audiostreamer.AudioProvider{}
	audioStreamer := audiostreamer.NewAudioStreamer(followers, interval, lookAhead, audioProvider)

	return followers, audioStreamer
}

func main() {
	followers, audioStreamer := initializeFixtures()
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	interval := 100 * time.Millisecond

	log("Subscribe Followers")
	cpURL := "127.0.0.1:9001/connections"
	follower.SubscribeFollower(followers, cpURL, ntpPort)

	log("Start NTP service")
	go clocksync.ExposeNTP(ntpPort, stop)

	log("Start AudioStreamLoop with sending interval " + interval.String())
	go audioStreamer.StreamAudioToAllLoop(interval, stop)

	fmt.Println("Continue? [y]es")
	var input string
	fmt.Scan(&input)
	if input == "y" {
		filePath := "./assets/test_audio.mp3"
		audioStreamer.AddToQueue(filePath)
	}

	<-sigChan
	log("Shutting down...")
	time.Sleep(time.Millisecond * 500)
	log("Exit.")
}
