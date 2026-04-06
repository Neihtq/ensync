package sourceprovider

import (
	"fmt"
	"time"

	"ensync/internal/grandmaster/navidrome"
)

type NaviDromeProvider struct {
	NaviDromeClient *navidrome.NavidromeClient
}

func NewNaviDromeProvider() *NaviDromeProvider {
	client := navidrome.NewNavidromeClient()
	provider := &NaviDromeProvider{
		NaviDromeClient: client,
	}
	go provider.checkHealth()

	return provider
}

func (provider *NaviDromeProvider) GetSource(trackIdentifier string) (*Decoder, error) {
	return nil, nil
}

func (provider *NaviDromeProvider) ListSongs() []string {
	return []string{}
}

func (provider *NaviDromeProvider) SearchSong(query string) []navidrome.Song {
	searchResult, err := provider.NaviDromeClient.Search(query)
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
		if err := provider.NaviDromeClient.Ping(); err != nil {
			fmt.Printf("Navidrome host is not reachable (attempt %d of %d): %v", errCount, maxAttempts, err)
			errCount++
		} else {
			errCount = 0
		}
		time.Sleep(5 * time.Second)
	}
	panic("Navidrome is offline!")
}
