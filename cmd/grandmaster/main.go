package main

import (
	"bufio"
	"os"
	"os/signal"
	"strings"
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
	followerServicePort = ":65535"
	ntpPort             = ":65534"
	webPort             = ":65533"

	// manualPeersFile lists follower IPs (one per line) to subscribe to directly,
	// bypassing mDNS discovery. Useful on networks that block multicast (e.g. phone
	// hotspots) or where the grandmaster otherwise cannot discover followers.
	manualPeersFile = ".config"
	// followerControlPort is the follower's control-plane port and endpoint that the
	// grandmaster POSTs to when subscribing. Must match the follower's cpPort/endpoint.
	followerControlPort = "65531"
	followerEndpoint    = "/connections"
)

func log(message string) {
	logging.Log(logPrefix, message)
}

func provideNavidromeProvider() *sourceprovider.NaviDromeProvider {
	return sourceprovider.NewNaviDromeProvider()
}

func provideFollowersRegistry() *follower.FollowersRegistry {
	heartbeatPort := ntpPort
	return follower.NewFollowersRegistry(heartbeatPort)
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

// readManualPeers loads follower IPs from manualPeersFile, one per line. Blank
// lines and lines beginning with '#' are ignored. A missing file is not an error;
// it simply means no manual peers are configured.
func readManualPeers(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var peers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		peers = append(peers, line)
	}

	return peers
}

// subscribeManualPeers keeps each configured follower IP registered, bypassing
// mDNS discovery. Each peer is watched in its own goroutine that (re)subscribes
// whenever the follower is absent from the registry — so a follower that drops off
// and is later evicted gets re-registered once it becomes reachable again.
func subscribeManualPeers(registry *follower.FollowersRegistry, peers []string, stop chan struct{}) {
	for _, ip := range peers {
		go func(ip string) {
			url := ip + ":" + followerControlPort + followerEndpoint
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				if !registry.IsRegistered(ip) {
					if err := follower.SubscribeFollower(registry, url, ntpPort); err != nil {
						log("Manual subscribe failed for " + url + ": " + err.Error() + " (retrying)")
					} else {
						log("Manually subscribed follower " + url)
					}
				}
				select {
				case <-stop:
					return
				case <-ticker.C:
				}
			}
		}(ip)
	}
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

	log("Start Follower service")
	followersRegistry := provideFollowersRegistry()
	followersRegistry.StartFollowerService(stop)

	log("Start ClockSync service")
	clockSyncService := provideClockSyncService()
	go clockSyncService.ExposeNTP(stop)

	provider := provideNavidromeProvider()
	log("Initialized Source Provider: Navidrome connector " + provider.Client.ApiVersion)

	log("Start AudioStreamLoop")
	trackQueue := provideTrackQueue()
	audioStreamer := provideStreamer(followersRegistry, provider, trackQueue)
	go audioStreamer.StreamAudioToAllLoop(stop)

	log("Start Discovery Service")
	discoveryService := provideDiscoveryService(followersRegistry)
	discoveryService.StartDiscovery()

	if peers := readManualPeers(manualPeersFile); len(peers) > 0 {
		log("Subscribing manual peers from " + manualPeersFile + ": " + strings.Join(peers, ", "))
		subscribeManualPeers(followersRegistry, peers, stop)
	}

	log("Start Web Server")
	webServer := provideWebserver(provider, followersRegistry, trackQueue)
	trackQueue.SetCallbackHook(webServer.BroadcastQueueState)
	followersRegistry.SetCallbackHook(webServer.BroadcastRegistry)
	go webServer.StartServer()

	<-sigChan
	log("Shutting down...")
	time.Sleep(time.Millisecond * 500)
	log("Exit.")
}
