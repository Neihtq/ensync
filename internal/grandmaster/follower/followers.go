// Package follower for Follower Service
package follower

import (
	"encoding/json"
	"net/http"
	"sync"

	"ensync/internal/grandmaster/logging"
)

const followsLogPrefix = "[Followers]"

type followersRequest struct {
	Address       string `json:"address"`
	HeartbeatPort string `json:"heartbeatPort"`
	AudioPort     string `json:"audioPort"`
}

type Followers struct {
	sync.RWMutex
	HeartbeatURLs []string
	AudioURLs     []string
}

func (s *Followers) CreateFollower(writer http.ResponseWriter, request *http.Request) {
	logging.Log(followsLogPrefix, "Received request for /followers")
	s.Lock()
	defer s.Unlock()

	var req followersRequest
	json.NewDecoder(request.Body).Decode(&req)
	addr := req.Address
	if addr == "" {
		panic("Provided address is empty!")
	}
	heartbeatURL := addr + ":" + req.HeartbeatPort
	audioURL := addr + ":" + req.AudioPort
	s.HeartbeatURLs = append(s.HeartbeatURLs, heartbeatURL)
	s.AudioURLs = append(s.AudioURLs, audioURL)

	logging.Log(followsLogPrefix, "Created Follower: "+addr)
}
