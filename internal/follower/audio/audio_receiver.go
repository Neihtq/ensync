package audio

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"ensync/internal/follower/middleware"
	"ensync/internal/follower/mirrorclock"
)

const headerSize = 8

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

	buffer := make([]byte, headerSize+3528)
	for {
		numBytes, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}

		timestamp := binary.BigEndian.Uint64(buffer[:headerSize])
		playAt := int64(timestamp)
		audio := buffer[headerSize:numBytes]

		audioStream.WriteToBuffer(audio, playAt)
	}
}

func checkPlayerError(player ErrorReporter, sleepInterval time.Duration) {
	for {
		if err := player.Err(); err != nil {
			fmt.Printf("OTO PLAYER FATAL ERROR: %v\n", err)
		}
		time.Sleep(sleepInterval)
	}
}

func LaunchAudioServer(
	port string,
	ipProvider middleware.IPProvider,
	clock *mirrorclock.MirrorClock,
	stop chan struct{},
) {
	audioStream := NewAudioStream(clock)
	context, player := NewPlayer(audioStream)
	fmt.Println("Context {}, Player {}", context, player)

	localIPAddress := ipProvider.GetIP().String()
	url := localIPAddress + ":" + port
	go expose(url, audioStream, stop)

	go checkPlayerError(player, 1*time.Second)
}
