// Package webservice implements the backend of the web app
package webservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ensync/internal/grandmaster/follower"
	"ensync/internal/grandmaster/sourceprovider"
)

type WebServer struct {
	Port              string
	SourceProvider    sourceprovider.SourceProvider
	FollowersRegistry *follower.FollowersRegistry
}

func NewWebServer(
	port string,
	sourceProvider sourceprovider.SourceProvider,
	followersRegistry *follower.FollowersRegistry,
) *WebServer {
	return &WebServer{
		Port:              port,
		SourceProvider:    sourceProvider,
		FollowersRegistry: followersRegistry,
	}
}

func (server *WebServer) StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	mux.HandleFunc("GET /songs", server.listSongs)
	mux.HandleFunc("GET /followers", server.listFollowers)

	fmt.Println("Webserver running on port", server.Port)
	http.ListenAndServe(server.Port, mux)
}

func (server *WebServer) listSongs(writer http.ResponseWriter, _ *http.Request) {
	songs := []string{"track1", "track2"}

	writer.Header().Set("Content-Type", "application/json")

	response := map[string][]string{
		"songs": songs,
	}
	json.NewEncoder(writer).Encode(response)
}

func (server *WebServer) listFollowers(writer http.ResponseWriter, _ *http.Request) {
	followerUrls := server.FollowersRegistry.GetAllFollowers()

	writer.Header().Set("Content-Type", "application/json")

	response := map[string][]string{
		"followerUrls": followerUrls,
	}
	json.NewEncoder(writer).Encode(response)
}
