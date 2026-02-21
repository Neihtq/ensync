package audiostreamer

import (
	"net"
	"testing"
)

func TestStreamAutoToAllLoop(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	serverAddr := conn.LocalAddr().String()
	received := make(chan string)
	go func() {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		received <- string(buf[:n])
	}()

	urls := []string{serverAddr}
	filePath := "../../../assets/test_file.mp3"
	StreamAudioToAll(filePath, urls)
}
