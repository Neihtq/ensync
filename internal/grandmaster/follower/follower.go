package follower

import (
	"log"
	"net"
	"sync"
)

type Follower struct {
	sync.RWMutex
	AudioURL string
	Conn     *net.UDPConn
}

func (f *Follower) InitConnection() {
	f.Lock()
	defer f.Unlock()
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
	f.Lock()
	defer f.Unlock()
	return f.Conn
}

func NewFollower(audioURL string) Follower {
	return Follower{
		AudioURL: audioURL,
	}
}
