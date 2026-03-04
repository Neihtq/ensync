// Package follower for Follower Service
package follower

import (
	"log"
	"net"
	"sync"
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
	AudioURL string
	Conn     *net.UDPConn
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

func NewFollower(audioURL string) Follower {
	return Follower{
		AudioURL: audioURL,
	}
}

type Followers struct {
	sync.RWMutex
	Followers []Follower
}

func NewFollowers() *Followers {
	return &Followers{
		Followers: []Follower{},
	}
}

func (followers *Followers) RegisterFollower(addr string) {
	followers.Lock()
	defer followers.Unlock()

	followers.Followers = append(followers.Followers, NewFollower(addr))
}
