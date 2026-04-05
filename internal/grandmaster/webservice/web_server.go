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
	fileServer := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fileServer)

	fmt.Println("Webserver running on port", server.Port)
	mux := http.NewServeMux()
	http.ListenAndServe(server.Port, mux)
}
