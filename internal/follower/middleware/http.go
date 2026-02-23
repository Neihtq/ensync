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

func Delete(baseURL string, param string) error {
	url := baseURL + "/" + param
	fmt.Printf("Calling %s\n", url)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	return nil
}
