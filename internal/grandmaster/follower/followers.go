// Package follower for Follower Service
package follower

import (
	"fmt"
	"log"
	"net"
	"strings"
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
	if f.Conn != nil {
		return
	}
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
	Followers map[string]Follower
}

func NewFollowers() *Followers {
	return &Followers{
		Followers: make(map[string]Follower),
	}
}

func (followers *Followers) RegisterFollower(ipAddress string, port string) {
	followers.Lock()
	defer followers.Unlock()

	fmt.Println("port", port)
	audioURL := ipAddress + ":" + strings.Trim(port, ":")

	if _, exists := followers.Followers[ipAddress]; !exists {
		fmt.Println("Registering new Follower", ipAddress)
		followers.Followers[ipAddress] = NewFollower(audioURL)
	}

	follower := followers.Followers[ipAddress]
	follower.InitConnection()
}
