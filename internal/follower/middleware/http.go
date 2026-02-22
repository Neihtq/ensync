// Package middleware contains code for connecting between servicees
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func Post(data map[string]string, url string) error {
	fmt.Printf("Calling %s\n", url)
	fmt.Printf("Payload %s\n", data)
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	return nil
}
