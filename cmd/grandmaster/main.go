package main

import (
	"fmt"

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
	audioStreamer := &audiostreamer.AudioStreamer{Subs: subscribers}

	return subscribers, publisher, audioStreamer
}

func main() {
	subscribers, publisher, audioStreamer := initializeFixtures()
	go subscription.SubscriptionService(subscribers, subscriptionServicePort)

	log("Start Heartbeat loop")
	go publisher.HeartbeatLoop()

	log("Start AudioStreamLoop")
	go audioStreamer.StreamAudioToAllLoop()

	fmt.Println("Continue? [y]es")
	var input string
	fmt.Scan(&input)
	if input == "y" {
		filePath := "./assets/test_audio.mp3"
		audioStreamer.AddToQueue(filePath)
	}
	for {
	}
}
