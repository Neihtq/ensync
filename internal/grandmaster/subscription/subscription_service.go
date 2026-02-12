package subscription

import (
	"net/http"

	"ensync/internal/grandmaster/logging"
)

const logPrefix = "[SubscriptionService]"

func log(message string) {
	logging.Log(logPrefix, message)
}

func SubscriptionService(subscribers *Subscribers, port string) {
	log("Starting SubscriptionService")
	mux := http.NewServeMux()

	log("Initialize '/subscribe' endpoint")
	mux.HandleFunc("POST /subscribe", subscribers.Subscribe)

	log("Listening on port " + port)
	http.ListenAndServe(port, mux)
}
