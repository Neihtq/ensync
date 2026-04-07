package controlplane

import (
	"bytes"
	"net"
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
	audioPort := ":4222"
	cp := NewControlPlaneService(mirrorClock, audioPort, stop)
	servicePort := ":42050"

	// act
	go cp.StartService(servicePort)

	close(stop)
}

func TestStartClockSync(t *testing.T) {
	// arrange
	mirrorClock := mirrorclock.NewMirrorClock()
	stop := make(chan struct{})
	audioPort := ":42051"
	cp := NewControlPlaneService(mirrorClock, audioPort, stop)

	port := ":42052"
	ipAddress := "127.0.0.1"
	address := ipAddress + port
	jsonBody := []byte(`{"address": "` + address + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/connections", bytes.NewBuffer(jsonBody))
	request.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	tcpAddr, _ := net.ResolveTCPAddr("tcp", address)
	tcpConn, _ := net.ListenTCP("tcp", tcpAddr)
	defer tcpConn.Close()

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
	expected := `{"address":"` + ipProvider.GetIP().String() + `","port":"` + cp.AudioPort + `"}`
	if !strings.Contains(writer.Body.String(), expected) {
		t.Errorf("Response body mismatch. Expected %s but got %s", expected, writer.Body.String())
	}
}
