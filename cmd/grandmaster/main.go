package main

import (
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

func main() {
	log("Initialize registry of Subscribers")
	subscribers := &subscription.Subscribers{}
	log("Initialize Heartbeat Publisher")
	publisher := &heartbeat.HeartbeatPublisher{Subs: subscribers}

	go subscription.SubscriptionService(subscribers, subscriptionServicePort)

	log("Start Heartbeat loop")
	publisher.HeartbeatLoop()
}
