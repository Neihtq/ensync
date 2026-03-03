package controlplane

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

func TestStartService(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})
	ntpPort := "4222"
	cp := NewControlPlaneService(mirrorClock, ntpPort, stop)
	servicePort := ":42050"

	// act
	go cp.StartService(servicePort)

	close(stop)
}

func TestStartClockSync(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})
	ntpPort := ":42051"
	cp := NewControlPlaneService(mirrorClock, ntpPort, stop)

	port := ":42052"
	ipAddress := "127.0.0.1"
	address := ipAddress + port
	jsonBody := []byte(`{"address": "` + address + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/connections", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	// act
	cp.StartClockSync(writer, request)

	// assert
	if cp.ClockSync == nil {
		t.Error("ControlPlane's Clock Sync should not be nil.")
	}
	if writer.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", writer.Code)
	}

	ipProvider := middleware.RealIPProvider{}
	outboundAddr := ipProvider.GetIP().String() + cp.ClockSyncPort
	expected := `{"address":"` + outboundAddr + `"}`
	if !strings.Contains(writer.Body.String(), expected) {
		t.Errorf("Response body mismatch. Expected %s but got %s", expected, writer.Body.String())
	}
}
