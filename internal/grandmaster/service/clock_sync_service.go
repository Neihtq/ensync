// Package service provides sclasses for all services in ensync
package service

import (
	"fmt"

	"ensync/internal/grandmaster/clocksync"
)

type ClockSyncService struct {
	port string
}

func NewClockSyncService(port string) *ClockSyncService {
	return &ClockSyncService{port: port}
}

func (service *ClockSyncService) Start(stop chan struct{}) {
	fmt.Println("Starting NTP Clock Sync Service")
	go clocksync.ExposeNTP(service.port, stop)
}
