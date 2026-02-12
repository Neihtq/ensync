// Package subscription for connection related logic
package subscription

import (
	"encoding/json"
	"net/http"
	"sync"

	"ensync/internal/grandmaster/logging"
)

const subscribersLogPrefix = "[Subscribers]"

type subscribeRequest struct {
	URL string `json:"url"`
}

type Subscribers struct {
	sync.RWMutex
	Urls []string
}

func (s *Subscribers) Subscribe(writer http.ResponseWriter, request *http.Request) {
	logging.Log(subscribersLogPrefix, "Received request for /subscribe")
	s.Lock()
	defer s.Unlock()

	var req subscribeRequest
	json.NewDecoder(request.Body).Decode(&req)
	url := req.URL
	s.Urls = append(s.Urls, url)

	logging.Log(subscribersLogPrefix, "Subscribed: "+url)
}
