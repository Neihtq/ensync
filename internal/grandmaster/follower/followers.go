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

type Follower struct {
	HeartbeatURL string
	AudioURL     string
}

func NewFollower(heartbeatURL string, audioURL string) Follower {
	return Follower{
		HeartbeatURL: heartbeatURL,
		AudioURL:     audioURL,
	}
}

type Followers struct {
	sync.RWMutex
	Followers map[string]Follower
}

func NewFollowers() *Followers {
	return &Followers{
		Followers: make(map[string]Follower),
	}
}

func (followers *Followers) AddFollower(writer http.ResponseWriter, request *http.Request) {
	logging.Log(followsLogPrefix, "Received request for /followers")
	followers.Lock()
	defer followers.Unlock()

	var req followersRequest
	json.NewDecoder(request.Body).Decode(&req)
	addr := req.Address
	if addr == "" {
		panic("Provided address is empty!")
	}

	followers.Followers[addr] = NewFollower(
		addr+":"+req.HeartbeatPort,
		addr+":"+req.AudioPort,
	)

	logging.Log(followsLogPrefix, "Created Follower: "+addr)
}
