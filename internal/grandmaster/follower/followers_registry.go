// Package follower for Follower Service
package follower

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const followsLogPrefix = "[FollowersRegistry]"

type addFollowersRequest struct {
	Address       string `json:"address"`
	HeartbeatPort string `json:"heartbeatPort"`
	AudioPort     string `json:"audioPort"`
}

type removeFollowersRequest struct {
	Address string `json:"address"`
}

type FollowersRegistry struct {
	sync.RWMutex
	Registry      map[string]*Follower
	HeartbeatPort string

	OnRegistryChanged func(followerUrls []string)
}

func NewFollowersRegistry(heartbeatPort string) *FollowersRegistry {
	return &FollowersRegistry{
		Registry:      make(map[string]*Follower),
		HeartbeatPort: heartbeatPort,
	}
}

func (registry *FollowersRegistry) RegisterFollower(ipAddress string, port string) {
	registry.Lock()

	audioURL := ipAddress + ":" + strings.Trim(port, ":")

	if _, exists := registry.Registry[ipAddress]; !exists {
		newFollower := NewFollower(audioURL)
		registry.Registry[ipAddress] = &newFollower
	} else {
		registry.Registry[ipAddress].SetAudioURL(audioURL)
	}
	registry.Unlock()

	callHook(registry)
}

func (registry *FollowersRegistry) GetAllFollowers() []string {
	registry.Lock()
	defer registry.Unlock()

	followerUrls := []string{}
	for url := range registry.Registry {
		followerUrls = append(followerUrls, url)
	}

	return followerUrls
}

func (registry *FollowersRegistry) StartHeartbeatService(stop chan struct{}) {
	addr, err := net.ResolveTCPAddr("tcp", registry.HeartbeatPort)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Heartbeat service listening on ", listener.Addr().String())

	for {
		select {
		case <-stop:
			return
		default:
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Error accepting conn:", err)
				continue
			}
			registry.HandleHeartbeat(conn)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (registry *FollowersRegistry) HandleHeartbeat(conn *net.TCPConn) {
	clientAddr := conn.RemoteAddr().String()
	ipAddress, _, err := net.SplitHostPort(clientAddr)
	if err != nil {
		fmt.Println("Error parsing remote address:", err)
		return
	}
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(10 * time.Second)

	if _, exists := registry.Registry[ipAddress]; !exists {
		newFollower := NewFollowerWithTCP(conn)
		registry.Lock()
		registry.Registry[ipAddress] = &newFollower
		registry.Unlock()
	} else {
		registry.Registry[ipAddress].SetTCPConn(conn)
	}

	logMessage("Follower joined Heartbeat service: " + ipAddress)
}

func (registry *FollowersRegistry) UnsubscribeFollower(addr string) {
	registry.Lock()

	if f, exists := registry.Registry[addr]; exists {
		if conn := f.TCPConn; conn != nil {
			conn.Close()
		}
	}
	delete(registry.Registry, addr)
	registry.Unlock()

	callHook(registry)
}

func (registry *FollowersRegistry) SetCallbackHook(
	onRegistryChanged func(followerUrls []string),
) {
	registry.OnRegistryChanged = onRegistryChanged
}

func callHook(registry *FollowersRegistry) {
	if registry.OnRegistryChanged != nil {
		followerUrls := registry.GetAllFollowers()
		registry.OnRegistryChanged(followerUrls)
	}
}

func (registry *FollowersRegistry) CheckHealthyFollowers(stop chan struct{}) {
	buf := make([]byte, 1)
	for {
		select {
		case <-stop:
			return
		default:
			registry.Lock()
			for addr, follower := range registry.Registry {
				conn := follower.GetTCPConn()
				if conn == nil {
					continue
				}
				_, err := conn.Read(buf)
				if err != nil {
					logMessage("Evicting unreachable Follower " + addr)
					registry.UnsubscribeFollower(addr)
				}
			}
			registry.Unlock()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (registry *FollowersRegistry) StartFollowerService(stop chan struct{}) {
	go registry.StartHeartbeatService(stop)
	go registry.CheckHealthyFollowers(stop)
}
