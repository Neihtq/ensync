// Package follower for Follower Service
package follower

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	"ensync/internal/grandmaster/logging"
)

const followsLogPrefix = "[Followers]"

type addFollowersRequest struct {
	Address       string `json:"address"`
	HeartbeatPort string `json:"heartbeatPort"`
	AudioPort     string `json:"audioPort"`
}

type removeFollowersRequest struct {
	Address string `json:"address"`
}

type Follower struct {
	HeartbeatURL string
	AudioURL     string
	Conn         *net.UDPConn
}

func (f *Follower) InitConnection() {
	addr, err := net.ResolveUDPAddr("udp", f.AudioURL)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	f.Conn = conn
}

func (f *Follower) GetConnection() *net.UDPConn {
	return f.Conn
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
	logging.Log(followsLogPrefix, "Received POST request for /followers")
	followers.Lock()
	defer followers.Unlock()

	var req addFollowersRequest
	json.NewDecoder(request.Body).Decode(&req)
	addr := req.Address
	heartbeatPort := req.HeartbeatPort
	audioPort := req.AudioPort
	if addr == "" || heartbeatPort == "" || audioPort == "" {
		http.Error(writer, "Address, Heartbeat Port, or Audio Port are missing or empty", http.StatusBadRequest)
	}

	followers.Followers[addr] = NewFollower(
		addr+":"+heartbeatPort,
		addr+":"+audioPort,
	)

	logging.Log(followsLogPrefix, "Created Follower: "+addr)
	writer.WriteHeader(http.StatusCreated)
}

func (followers *Followers) RemoveFollower(writer http.ResponseWriter, request *http.Request) {
	logging.Log(followsLogPrefix, "Received DELETE request for /followers")
	followers.Lock()
	defer followers.Unlock()

	addr := request.PathValue("address")

	if addr == "" {
		http.Error(writer, "Provided address is empty", http.StatusBadRequest)
		return
	}

	if _, exists := followers.Followers[addr]; exists {
		delete(followers.Followers, addr)
		logging.Log(followsLogPrefix, "Removed Follower: "+addr)
	}

	writer.WriteHeader(http.StatusNoContent)
}
