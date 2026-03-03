// Package controlplane implements the control plane of the Follower
package controlplane

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ensync/internal/follower/clocksync"
	"ensync/internal/follower/mirrorclock"
)

type ConnectionsPOSTRequest struct {
	IPAddress string `json:"address"`
	Port      string `json:"port"`
}

type ControlPlaneService struct {
	Clock     *mirrorclock.MirrorClock
	ClockSync *clocksync.ClockSync
	stop      chan struct{}
}

func NewControlPlaneService(clock *mirrorclock.MirrorClock, stop chan struct{}) *ControlPlaneService {
	return &ControlPlaneService{
		Clock: clock,
		stop:  stop,
	}
}

func (cp *ControlPlaneService) StartClockSync(writer http.ResponseWriter, request *http.Request) {
	var req ConnectionsPOSTRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}
	ipAddress := req.IPAddress
	port := req.Port
	fmt.Println("IpAddress, Port", ipAddress, port)
	if ipAddress == "" || port == "" {
		return
	}

	fullAddress := ipAddress + ":" + port
	cp.ClockSync = clocksync.NewClockSync(cp.Clock, fullAddress)
	go cp.ClockSync.RunClockSync(cp.stop)
	fmt.Println("Launched ClockSync service to ", ipAddress)
}

func (cp *ControlPlaneService) StartService(port string) {
	fmt.Println("Starting ControlPlane")
	mux := http.NewServeMux()

	mux.HandleFunc("POST /connections", cp.StartClockSync)

	http.ListenAndServe(port, mux)
}
