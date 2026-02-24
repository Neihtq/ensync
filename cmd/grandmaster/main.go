package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/grandmaster/audiostreamer"
	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/heartbeat"
	"ensync/internal/grandmaster/logging"
)

const (
	logPrefix           = "[Main]"
	followerServicePort = ":8080"
)

func log(message string) {
	logging.Log(logPrefix, message)
}

func initializeFixtures() (*follower.Followers, *heartbeat.HeartbeatPublisher, *audiostreamer.AudioStreamer) {
	log("Initialize registry of Subscribers")
	followers := follower.NewFollowers()
	log("Initialize Heartbeat Publisher")
	publisher := &heartbeat.HeartbeatPublisher{Followers: followers}

	log("Initialize AudioStreamer")
	interval := 10 * time.Millisecond
	lookAhead := (200 * time.Millisecond).Nanoseconds()
	audioProvider := &audiostreamer.AudioProvider{}
	audioStreamer := audiostreamer.NewAudioStreamer(followers, interval, lookAhead, audioProvider)

	return followers, publisher, audioStreamer
}

func main() {
	followers, publisher, audioStreamer := initializeFixtures()
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go follower.FollowerService(followers, followerServicePort)

	interval := 100 * time.Millisecond

	log("Start Heartbeat loop")
	go publisher.HeartbeatLoop(interval, stop)

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
