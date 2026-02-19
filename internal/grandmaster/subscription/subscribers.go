// Package subscription for connection related logic
package subscription

import (
	"encoding/json"
	"net/http"
	"sync"

	"ensync/internal/grandmaster/logging"
)

const subscribersLogPrefix = "[Subscribers]"

type subscribeRequest struct {
	Address       string `json:"address"`
	HeartbeatPort string `json:"heartbeatPort"`
	AudioPort     string `json:"audioPort"`
}

type Subscribers struct {
	sync.RWMutex
	HeartbeatURLs []string
	AudioURLs     []string
}

func (s *Subscribers) Subscribe(writer http.ResponseWriter, request *http.Request) {
	logging.Log(subscribersLogPrefix, "Received request for /subscribe")
	s.Lock()
	defer s.Unlock()

	var req subscribeRequest
	json.NewDecoder(request.Body).Decode(&req)
	addr := req.Address
	if addr == "" {
		panic("Provided address is empty!")
	}
	heartbeatURL := addr + ":" + req.HeartbeatPort
	audioURL := addr + ":" + req.AudioPort
	s.HeartbeatURLs = append(s.HeartbeatURLs, heartbeatURL)
	s.AudioURLs = append(s.AudioURLs, audioURL)

	logging.Log(subscribersLogPrefix, "Subscribed: "+addr)
}
