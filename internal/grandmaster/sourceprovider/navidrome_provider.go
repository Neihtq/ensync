package sourceprovider

import (
	"fmt"
	"time"

	"ensync/internal/grandmaster/navidrome"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/vorbis"
)

type NaviDromeProvider struct {
	Client *navidrome.NavidromeClient
}

func NewNaviDromeProvider() *NaviDromeProvider {
	client := navidrome.NewNavidromeClient()
	provider := &NaviDromeProvider{
		Client: client,
	}
	go provider.checkHealth()

	return provider
}

func (provider *NaviDromeProvider) GetSource(trackIdentifier string) (*Decoder, error) {
	stream, contentType, err := provider.Client.GetStream(trackIdentifier)
	if err != nil {
		return nil, fmt.Errorf("Error fetching stream %w", err)
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format
	switch contentType {
	case "audio/flac", "application/x-flac":
		streamer, format, err = flac.Decode(stream)
	case "audio/mp4", "audio/acc":
		streamer, format, err = vorbis.Decode(stream)
	default:
		streamer, format, err = mp3.Decode(stream)
	}

	if err != nil {
		stream.Close()
		return nil, fmt.Errorf("decode error for %s %w", contentType, err)
	}

	decoder := &Decoder{
		Streamer:   streamer,
		Closer:     stream,
		SampleRate: sampleRate,
		Channels:   2,
	}

	Resample(format, streamer, decoder)
	return decoder, nil
}

func (provider *NaviDromeProvider) ListSongs() []string {
	return []string{}
}

func (provider *NaviDromeProvider) SearchSong(query string) []navidrome.Song {
	searchResult, err := provider.Client.Search(query)
	if err != nil {
		fmt.Println("Error calling /search3", err)
		return []navidrome.Song{}
	}

	if searchResult == nil {
		return []navidrome.Song{}
	}

	return searchResult.Song
}

func (provider *NaviDromeProvider) checkHealth() {
	errCount := 0
	maxAttempts := 10
	for errCount < maxAttempts {
		if err := provider.Client.Ping(); err != nil {
			fmt.Printf("Navidrome host is not reachable (attempt %d of %d): %v", errCount, maxAttempts, err)
			errCount++
		} else {
			errCount = 0
		}
		time.Sleep(5 * time.Second)
	}
	panic("Navidrome is offline!")
}

func (provider *NaviDromeProvider) GetSong(trackID string) (*navidrome.Song, error) {
	song, err := provider.Client.GetSong(trackID)
	if err != nil {
		return nil, fmt.Errorf("Error fetching song %w", err)
	}

	return song, nil
}

func (provider *NaviDromeProvider) GetTitle(trackIdentifier string) string {
	song, err := provider.GetSong(trackIdentifier)
	if err != nil {
		fmt.Println("Error getting song title", err)
		return ""
	}
	return song.Title + " - " + song.Artist
}
