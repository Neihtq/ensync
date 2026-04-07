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
	TCPConn  *net.TCPConn
}

func (f *Follower) GetTCPConn() *net.TCPConn {
	f.Lock()
	defer f.Unlock()

	return f.TCPConn
}

func (f *Follower) SetTCPConn(conn *net.TCPConn) {
	f.Lock()
	defer f.Unlock()
	f.TCPConn = conn
}

func (f *Follower) SetAudioURL(url string) {
	f.Lock()
	defer f.Unlock()

	f.AudioURL = url
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

func NewFollowerWithTCP(conn *net.TCPConn) Follower {
	return Follower{
		TCPConn: conn,
	}
}
