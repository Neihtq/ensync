// Package follower for Follower Service
package follower

import (
	"strings"
	"sync"
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
	Registry map[string]*Follower
}

func NewFollowersRegistry() *FollowersRegistry {
	return &FollowersRegistry{
		Registry: make(map[string]*Follower),
	}
}

func (registry *FollowersRegistry) RegisterFollower(ipAddress string, port string) {
	registry.Lock()
	defer registry.Unlock()

	audioURL := ipAddress + ":" + strings.Trim(port, ":")

	if _, exists := registry.Registry[ipAddress]; !exists {
		newFollower := NewFollower(audioURL)
		registry.Registry[ipAddress] = &newFollower
	}
}
