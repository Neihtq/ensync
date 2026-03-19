package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ensync/internal/grandmaster/audiostreamer"
	"ensync/internal/grandmaster/clocksync"
	"ensync/internal/grandmaster/discovery"
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

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Failed fetching outbound IP: ", err.Error())
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {
	followers, audioStreamer := initializeFixtures()
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	interval := 100 * time.Millisecond

	log("Start NTP service")
	go clocksync.ExposeNTP(ntpPort, stop)

	log("Start AudioStreamLoop with sending interval " + interval.String())
	go audioStreamer.StreamAudioToAllLoop(interval, stop)

	outboundIP := getOutboundIP()
	log("Start Discovery Lobby with IP:\n" + outboundIP)
	lobby := discovery.NewDiscoveryLobby(followers, stop)
	go lobby.OpenLobby(ntpPort)

	fmt.Println("Transfer visitors to followers? [y]es")
	var input string
	fmt.Scan(&input)
	lobby.TransferVisitorsToFollowers(ntpPort)

	fmt.Println("Continue? [y]es")
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
