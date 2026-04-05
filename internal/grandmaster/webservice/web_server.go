// Package webservice implements the backend of the web app
package webservice

import (
	"fmt"
	"net/http"
)

type WebServer struct {
	Port string
}

func NewWebServer(port string) *WebServer {
	return &WebServer{Port: port}
}

func (server *WebServer) StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	fmt.Println("Webserver running on port", server.Port)
	http.ListenAndServe(server.Port, mux)
}
