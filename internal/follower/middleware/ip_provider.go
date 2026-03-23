package middleware

import (
	"ensync/internal/common/netutil"
	"net"
)


type IPProvider interface {
	GetIP() net.IP
}

type RealIPProvider struct{}

func (r RealIPProvider) GetIP() net.IP {
	return netutil.GetOutboundIP()
}

type MockIPProvider struct {
	FakeIP net.IP
}

func (m MockIPProvider) GetIP() net.IP {
	return m.FakeIP
}

type NTPAddressProvider struct{}

func (n NTPAddressProvider) GetAddress(port string) string {
	ipAddr := netutil.GetOutboundIP().String()
	return ipAddr + port
}

func (n NTPAddressProvider) BuildAddress(url string, port string) string {
	return url + port
}
