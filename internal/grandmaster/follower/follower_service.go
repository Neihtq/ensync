package follower

import (
	"net/http"

	"ensync/internal/grandmaster/logging"
)

const logPrefix = "[FollowerService]"

func log(message string) {
	logging.Log(logPrefix, message)
}

func FollowerService(followers *Followers, port string) {
	log("Starting FollowerService")
	mux := http.NewServeMux()

	log("Initialize '/followers' endpoint")
	mux.HandleFunc("POST /followers", followers.AddFollower)

	log("Initialize '/followers' endpoint")
	mux.HandleFunc("DELETE /followers/{address}", followers.RemoveFollower)

	log("Listening on port " + port)
	http.ListenAndServe(port, mux)
}
