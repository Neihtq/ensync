package follower

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"ensync/internal/grandmaster/logging"
)

const logPrefix = "[FollowerService]"

func logMessage(message string) {
	logging.Log(logPrefix, message)
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func SubscribeFollower(followers *Followers, url string, ntpPort string) error {
	ipAddr := getOutboundIP()
	addr := ipAddr.String() + ":" + strings.Trim(ntpPort, ":")

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
	logMessage("Register follower " + result.Address)
	followers.RegisterFollower(result.Address, result.Port)

	return nil
}
