package controlplane

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"ensync/internal/follower/mirrorclock"
)

func TestStartService(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})
	cp := NewControlPlaneService(mirrorClock, stop)
	port := ":42050"

	// act
	go cp.StartService(port)

	close(stop)
}

func TestStartClockSync(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})
	cp := NewControlPlaneService(mirrorClock, stop)

	port := "42051"
	address := "127.0.0.1"
	jsonBody := []byte(`{"address": "` + address + `", "port": "` + port + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/connections", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	// act
	cp.StartClockSync(writer, request)

	// arrange
	if cp.ClockSync == nil {
		t.Error("ControlPlane's Clock Sync should not be nil.")
	}
}
