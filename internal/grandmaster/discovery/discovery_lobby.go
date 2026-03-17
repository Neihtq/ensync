package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type VisitorPOSTRequest struct {
	Address string `json:"address"`
	Port    string `json:"port"`
}

type DiscoveryLobby struct {
	stop     chan struct{}
	visitors map[string]string
}

func NewDiscoveryLobby(stop chan struct{}) *DiscoveryLobby {
	return &DiscoveryLobby{
		stop:     stop,
		visitors: make(map[string]string),
	}
}

func (dl *DiscoveryLobby) JoinLobby(writer http.ResponseWriter, request *http.Request) {
	var req VisitorPOSTRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		fmt.Println("invalid JSON")
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		fmt.Println("Missing Address")
		http.Error(writer, "Missing address", http.StatusBadRequest)
		return
	}
	dl.visitors[req.Address] = req.Port

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
}

func (dl *DiscoveryLobby) OpenLobby(port string) {
	fmt.Println("Open Lobby")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /visitor", dl.JoinLobby)
}
