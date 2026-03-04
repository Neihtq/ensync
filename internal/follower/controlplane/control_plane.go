// Package controlplane implements the control plane of the Follower
package controlplane

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ensync/internal/follower/clocksync"
	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

type ConnectionsPOSTRequest struct {
	Address string `json:"address"`
}

type ControlPlaneService struct {
	Clock         *mirrorclock.MirrorClock
	ClockSync     *clocksync.ClockSync
	ClockSyncPort string
	stop          chan struct{}
}

func NewControlPlaneService(clock *mirrorclock.MirrorClock, ntpPort string, stop chan struct{}) *ControlPlaneService {
	return &ControlPlaneService{
		Clock:         clock,
		ClockSyncPort: ntpPort,
		stop:          stop,
	}
}

func (cp *ControlPlaneService) StartClockSync(writer http.ResponseWriter, request *http.Request) {
	var req ConnectionsPOSTRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		fmt.Println("invalid JSON")
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}
	address := req.Address
	if address == "" {
		fmt.Println("Missing Address")
		http.Error(writer, "Missing address", http.StatusBadRequest)
		return
	}

	cp.ClockSync = clocksync.NewClockSync(cp.Clock, address)
	go cp.ClockSync.RunClockSync(cp.stop)
	fmt.Println("Launched ClockSync service to ", address)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)

	ipProvider := middleware.RealIPProvider{}
	outboundAddr := ipProvider.GetIP().String() + cp.ClockSyncPort
	response := map[string]string{"address": outboundAddr}
	json.NewEncoder(writer).Encode(response)
}

func (cp *ControlPlaneService) StartService(port string) {
	fmt.Println("Starting ControlPlane")
	mux := http.NewServeMux()

	mux.HandleFunc("POST /connections", cp.StartClockSync)

	http.ListenAndServe(port, mux)
}
