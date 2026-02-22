package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/grandmaster/audiostreamer"
	"ensync/internal/grandmaster/heartbeat"
	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/subscription"
)

const (
	logPrefix               = "[Main]"
	subscriptionServicePort = ":8080"
)

func log(message string) {
	logging.Log(logPrefix, message)
}

func initializeFixtures() (*subscription.Subscribers, *heartbeat.HeartbeatPublisher, *audiostreamer.AudioStreamer) {
	log("Initialize registry of Subscribers")
	subscribers := &subscription.Subscribers{}
	log("Initialize Heartbeat Publisher")
	publisher := &heartbeat.HeartbeatPublisher{Subs: subscribers}

	log("Initialize AudioStreamer")
	interval := 10 * time.Millisecond
	audioProvider := &audiostreamer.AudioProvider{}
	audioStreamer := audiostreamer.NewAudioStreamer(subscribers, interval, audioProvider)

	return subscribers, publisher, audioStreamer
}

func main() {
	subscribers, publisher, audioStreamer := initializeFixtures()
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go subscription.SubscriptionService(subscribers, subscriptionServicePort)

	log("Start Heartbeat loop")
	go publisher.HeartbeatLoop()

	interval := 100 * time.Millisecond
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
