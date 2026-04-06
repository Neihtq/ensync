// Package webservice implements the backend of the web app
package webservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/queue"
	"ensync/internal/grandmaster/sourceprovider"
)

type PushTrackRequest struct {
	TrackIdentifier string `json:"trackId"`
}

type WebServer struct {
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
	mux.HandleFunc("GET /queue", server.GetQueue)

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

func (server *WebServer) GetQueue(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	queue := server.TrackQueue.GetAllItems()

	response := map[string][]string{
		"tracks": queue,
	}
	json.NewEncoder(writer).Encode(response)
}
