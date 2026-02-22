package audio

import (
	"fmt"
	"net"
	"os"
	"time"

	"ensync/internal/follower/middleware"

	"github.com/ebitengine/oto/v3"
)

func expose(
	url string,
	audioStream *AudioStream,
	stop chan struct{},
) {
	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		fmt.Println("Error resolving address for audio server:\n", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting listener for audio server:\n", err)
	}
	defer conn.Close()

	fmt.Printf("Audio server listening on %s\n", url)

	go func() {
		<-stop
		conn.Close()
	}()

	buffer := make([]byte, 3528)
	for {
		numBytes, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}

		audioStream.WriteToBuffer(buffer[:numBytes])
	}
}

func checkPlayerError(player *oto.Player) {
	for {
		if err := player.Err(); err != nil {
			fmt.Printf("OTO PLAYER FATAL ERROR: %v\n", err)
		}
		time.Sleep(1 * time.Second)
	}
}

func LaunchAudioServer(port string, ipProvider middleware.IPProvider, stop chan struct{}) {
	audioStream := NewAudioStream()
	context, player := NewPlayer(audioStream)
	fmt.Println("Context {}, Player {}", context, player)

	localIPAddress := ipProvider.GetIP().String()
	url := localIPAddress + ":" + port
	go expose(url, audioStream, stop)

	go checkPlayerError(player)
}
