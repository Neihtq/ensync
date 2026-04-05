package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ensync/internal/grandmaster/follower"
)

type VisitorPOSTRequest struct {
	Address  string `json:"address"`
	Port     string `json:"port"`
	Endpoint string `json:"endpoint"`
}

type DiscoveryLobby struct {
	stop      chan struct{}
	visitors  map[string]string
	Followers *follower.FollowersRegistry
}

func (dl *DiscoveryLobby) TransferVisitorsToFollowers(ntpPort string) {
	for ipAddr, suffix := range dl.visitors {
		addr := ipAddr + suffix
		fmt.Println("Subscribe", addr)
		err := follower.SubscribeFollower(dl.Followers, addr, ntpPort)
		if err != nil {
			fmt.Println("Failed subscribing ", err.Error())
		}
	}
}

func NewDiscoveryLobby(followers *follower.FollowersRegistry, stop chan struct{}) *DiscoveryLobby {
	return &DiscoveryLobby{
		stop:      stop,
		visitors:  make(map[string]string),
		Followers: followers,
	}
}

func (dl *DiscoveryLobby) JoinLobby(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Received request")
	var req VisitorPOSTRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		fmt.Println("invalid JSON")
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	port := strings.Trim(req.Port, ":")
	endpoint := req.Endpoint
	dl.visitors[req.Address] = ":" + port + endpoint
	fmt.Println("Added visitor", req.Address)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
}

func (dl *DiscoveryLobby) OpenLobby(port string) {
	fmt.Println("Open Lobby")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /visitors", dl.JoinLobby)

	http.ListenAndServe(port, mux)
}
