package follower

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ensync/internal/common/netutil"
	"ensync/internal/grandmaster/logging"
)

const logPrefix = "[FollowerService]"

func logMessage(message string) {
	logging.Log(logPrefix, message)
}

func SubscribeFollower(followersRegistry *FollowersRegistry, url string, ntpPort string) error {
	ipAddr := netutil.GetOutboundIP()
	addr := ipAddr.String() + ":" + strings.Trim(ntpPort, ":")
	fmt.Println("Subscribing ", addr)
	data := map[string]string{"address": addr}
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post("http://"+url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	var result struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("server returned invalid JSON")
	}
	logMessage("Registering follower " + result.Address)
	followersRegistry.RegisterFollower(result.Address, result.Port)

	return nil
}
