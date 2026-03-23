package netutil

import (
	"errors"
	"net"
	"testing"
	"time"
)

type mockConn struct {
	localAddr net.Addr
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return m.localAddr }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestGetOutboundIP_InternetSuccess(t *testing.T) {
	expectedIP := net.ParseIP("192.168.1.100")
	defaultDial := dialFunc

	dialFunc = func(network, address string) (net.Conn, error) {
		return &mockConn{
			localAddr: &net.UDPAddr{IP: expectedIP, Port: 12345},
		}, nil
	}
	defer func() { dialFunc = defaultDial }()

	ip := GetOutboundIP()
	if !ip.Equal(expectedIP) {
		t.Errorf("Expected %v, got %v", expectedIP, ip)
	}
}

func TestGetOutboundIP_OfflineFallbackSuccess(t *testing.T) {
	expectedIP := net.ParseIP("10.0.0.50")
	defaultDial := dialFunc
	defaultInterface := interfaceFunc

	dialFunc = func(network, address string) (net.Conn, error) {
		return nil, errors.New("network unreachable")
	}

	interfaceFunc = func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)}, // Loopback, should skip
			&net.IPNet{IP: expectedIP, Mask: net.CIDRMask(24, 32)},              // Valid IPv4
		}, nil
	}

	defer func() {
		dialFunc = defaultDial
		interfaceFunc = defaultInterface
	}()

	ip := GetOutboundIP()
	if !ip.Equal(expectedIP) {
		t.Errorf("Expected offline fallback %v, got %v", expectedIP, ip)
	}
}

func TestGetOutboundIP_FatalFailure(t *testing.T) {
	defaultDial := dialFunc
	defaultInterface := interfaceFunc
	defaultFatal := fatalFunc

	dialFunc = func(network, address string) (net.Conn, error) {
		return nil, errors.New("network unreachable")
	}

	interfaceFunc = func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)}, // Only loopback
			&net.IPNet{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)},    // IPv6 Loopback
		}, nil
	}

	fatalCalled := false
	fatalFunc = func(v ...any) { fatalCalled = true }

	defer func() {
		dialFunc = defaultDial
		interfaceFunc = defaultInterface
		fatalFunc = defaultFatal
	}()

	ip := GetOutboundIP()
	if ip != nil {
		t.Errorf("Expected nil IP gracefully returning, got %v", ip)
	}
	if !fatalCalled {
		t.Errorf("Expected fatalFunc to be called when no valid IPs exist")
	}
}
