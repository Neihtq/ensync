// Package audiostreamer implements streaming capabilities of audio data
package audiostreamer

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"ensync/internal/grandmaster/logging"
	"ensync/internal/grandmaster/subscription"

	"github.com/gammazero/deque"
	"github.com/hajimehoshi/go-mp3"
)

const (
	interval  = 10 * time.Millisecond
	logPrefix = "[AudioStreamer]"
)

type AudioStreamer struct {
	mu    sync.Mutex
	Subs  *subscription.Subscribers
	Queue deque.Deque[string] // List of tracks
}

func NewAudioStreamer(subs *subscription.Subscribers) *AudioStreamer {
	return &AudioStreamer{Subs: subs}
}

func (streamer *AudioStreamer) AddToQueue(filePath string) {
	streamer.mu.Lock()
	defer streamer.mu.Unlock()
	streamer.Queue.PushBack(filePath)
}

func (streamer *AudioStreamer) StreamAudioToAllLoop() {
	for {
		streamer.mu.Lock()
		if streamer.Queue.Len() == 0 {
			streamer.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		filePath := streamer.Queue.PopFront()
		streamer.mu.Unlock()
		logging.Log(logPrefix, "Stream "+filePath)
		StreamAudioToAll(filePath, streamer.Subs.AudioURLs)
	}
}

func streamAudioToFollower(buffer []byte, url string) {
	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	conn.Write(buffer)
}

func StreamAudioToAll(filePath string, urls []string) {
	audioSource, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer audioSource.Close()

	decodedMp3, err := mp3.NewDecoder(audioSource)
	fmt.Println("MP3 Sample Rate: {}", decodedMp3.SampleRate())
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}

	ticker := time.NewTicker(interval)
	buffer := make([]byte, 3528)

	for range ticker.C {
		n, err := decodedMp3.Read(buffer)
		if n == 0 || err != nil {
			logging.Log(logPrefix, "Exiting play loop: n="+strconv.Itoa(n)+" err="+err.Error())
			break
		}

		for _, url := range urls {
			streamAudioToFollower(buffer[:n], url)
		}
	}
}
