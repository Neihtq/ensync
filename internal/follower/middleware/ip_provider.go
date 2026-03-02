package middleware

import "net"

func getOutboundIP(targetAddr string) net.IP {
	conn, err := net.Dial("udp", targetAddr)
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

type IPProvider interface {
	GetIP() net.IP
}

type RealIPProvider struct{}

func (r RealIPProvider) GetIP() net.IP {
	return getOutboundIP("8.8.8.8:80")
}

type MockIPProvider struct {
	FakeIP net.IP
}

func (m MockIPProvider) GetIP() net.IP {
	return m.FakeIP
}

type NTPAddressProvider struct{}

func (n NTPAddressProvider) GetAddress(port string) string {
	ipAddr := getOutboundIP("8.8.8.8:80").String()
	return ipAddr + port
}

func (n NTPAddressProvider) BuildAddress(url string, port string) string {
	return url + port
}
