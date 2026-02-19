// package audiostreamer implements streaming capabilities of audio data
package audiostreamer

import (
	"log"
	"net"
	"os"
	"time"

	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/subscription"
)

const (
	interval  = 20 * time.Millisecond
	logPrefix = "[AudioStreamer]"
)

type AudioStreamer struct {
	Subs *subscription.Subscribers
}

func streamAudioToFollower(buffer []byte, url string) {
	logging.Log(logPrefix, "Send audio")
	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
	}
	logging.Log(logPrefix, "Send audio packaget to address: "+addr.String())

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conn.Write(buffer)
}

func StreamAudioToAll(buffer []byte, urls []string) {
	for _, url := range urls {
		streamAudioToFollower(buffer, url)
	}
}

func (streamer *AudioStreamer) StreamAudioToAllLoop(filePath string) {
	audioSource, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer audioSource.Close()

	ticker := time.NewTicker(interval)
	buffer := make([]byte, 3528)

	for range ticker.C {
		n, err := audioSource.Read(buffer)
		if n == 0 || err != nil {
			break
		}

		StreamAudioToAll(buffer[:n], streamer.Subs.AudioURLs)
	}
}
