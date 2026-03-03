package follower

import (
	"net/http"

	"ensync/internal/grandmaster/logging"
)

const logPrefix = "[FollowerService]"

func logMessage(message string) {
	logging.Log(logPrefix, message)
}

func FollowerService(followers *Followers, port string) {
	logMessage("Starting FollowerService")
	mux := http.NewServeMux()

	logMessage("Initialize '/followers' endpoint")
	mux.HandleFunc("POST /followers", followers.AddFollower)

	logMessage("Initialize '/followers' endpoint")
	mux.HandleFunc("DELETE /followers/{address}", followers.RemoveFollower)

	logMessage("Listening on port " + port)
	http.ListenAndServe(port, mux)
}
