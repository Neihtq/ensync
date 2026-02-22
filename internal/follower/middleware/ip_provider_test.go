package middleware

import (
	"net"
	"testing"
)

func TestIPProvider(t *testing.T) {
	// arrange
	ipProvider := RealIPProvider{}

	// act
	ipAddress := ipProvider.GetIP()

	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()

	expected := conn.LocalAddr().(*net.UDPAddr).IP
	if ipAddress.String() != expected.String() {
		t.Fatalf("IP Provider failed: Expected %s but got %s", expected.String(), expected.String())
	}
}
