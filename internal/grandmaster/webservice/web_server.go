// Package webservice implements the backend of the web app
package webservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/queue"
	"ensync/internal/grandmaster/sourceprovider"
)

type PushTrackRequest struct {
	TrackIdentifier string `json:"trackId"`
}

type QueueState struct {
	NowPlaying string   `json:"nowPlaying"`
	QueueItems []string `json:"queueItems"`
}

type WebServer struct {
	mu                sync.Mutex
	connections       []chan QueueState
	Port              string
	SourceProvider    sourceprovider.SourceProvider
	FollowersRegistry *follower.FollowersRegistry
	TrackQueue        *queue.TrackQueue
}

func NewWebServer(
	port string,
	sourceProvider sourceprovider.SourceProvider,
	followersRegistry *follower.FollowersRegistry,
	trackQueue *queue.TrackQueue,
) *WebServer {
	return &WebServer{
		Port:              port,
		SourceProvider:    sourceProvider,
		FollowersRegistry: followersRegistry,
		TrackQueue:        trackQueue,
	}
}

func (server *WebServer) StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	mux.HandleFunc("GET /songs", server.ListSongs)
	mux.HandleFunc("GET /followers", server.ListFollowers)
	mux.HandleFunc("POST /tracks", server.PushTrack)
	mux.HandleFunc("GET /queue", server.StreamQueue)

	fmt.Println("Webserver running on port", server.Port)
	http.ListenAndServe(server.Port, mux)
}

func (server *WebServer) ListSongs(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	songs := server.SourceProvider.ListSongs()

	response := map[string][]string{
		"songs": songs,
	}
	json.NewEncoder(writer).Encode(response)
}

func (server *WebServer) ListFollowers(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	followerUrls := server.FollowersRegistry.GetAllFollowers()

	response := map[string][]string{
		"followerUrls": followerUrls,
	}
	json.NewEncoder(writer).Encode(response)
}

func (server *WebServer) PushTrack(writer http.ResponseWriter, request *http.Request) {
	var data PushTrackRequest
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		fmt.Println("PusTrackRequest: Invalid JSON")
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	trackIdentifier := data.TrackIdentifier
	server.TrackQueue.PushBack(trackIdentifier)
	writer.WriteHeader(http.StatusCreated)
}

func (server *WebServer) BroadcastQueueState(nowPlaying string, queueItems []string) {
	server.mu.Lock()
	defer server.mu.Unlock()

	state := QueueState{
		NowPlaying: nowPlaying,
		QueueItems: queueItems,
	}

	for _, ch := range server.connections {
		select {
		case ch <- state:
		default:
		}
	}
}

func (server *WebServer) StreamQueue(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming not supported!", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan QueueState, 1)

	server.mu.Lock()
	server.connections = append(server.connections, messageChan)
	server.mu.Unlock()

	server.TrackQueue.CallHook()
	for {
		select {
		case <-request.Context().Done():
			fmt.Println("Client disconnected")
			server.mu.Lock()
			for i, ch := range server.connections {
				if ch == messageChan {
					server.connections = append(server.connections[:i], server.connections[i+1:]...)
					break
				}
			}
			server.mu.Unlock()
			return
		case queueState := <-messageChan:
			payload, err := json.Marshal(queueState)
			if err != nil {
				fmt.Println("Error marshalling JSON:", err)
				continue
			}
			fmt.Fprintf(writer, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
