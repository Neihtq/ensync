// Package navidrome implements interfaces to communicate with the Subsonic API
// Subsonic API refs: https://www.subsonic.org/pages/api.jsp
package navidrome

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type NavidromeClient struct {
	BaseURL    string
	ApiVersion string
	ClientName string

	HttpClient *http.Client
}

func NewNavidromeClient() *NavidromeClient {
	serverURL := os.Getenv("NAVIDROME_URL")
	if serverURL == "" {
		panic("Server URL not set!")
	}
	return &NavidromeClient{
		BaseURL:    serverURL,
		ApiVersion: "1.16.1",
		ClientName: "NSYNC",
		HttpClient: &http.Client{},
	}
}

func generateSalt(numBytes int) string {
	bytes := make([]byte, numBytes)
	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

func (client *NavidromeClient) buildauthParms() url.Values {
	username, password := getCredentials()

	salt := generateSalt(8)
	saltedPassword := password + salt

	hash := md5.Sum([]byte(saltedPassword))
	authenticationToken := hex.EncodeToString(hash[:])

	params := url.Values{}
	params.Set("u", username)
	params.Set("t", authenticationToken)
	params.Set("s", salt)
	params.Set("v", client.ApiVersion)
	params.Set("c", client.ClientName)
	params.Set("f", "json")

	return params
}

func (client *NavidromeClient) Ping() error {
	_, err := client.callGet("ping", nil)
	if err != nil {
		return err
	}

	return nil
}

func (client *NavidromeClient) Search(query string) (*SearchResult3, error) {
	resultLimit := "20"
	params := url.Values{}
	params.Set("query", query)
	params.Set("artistCount", resultLimit)
	params.Set("albumCount", resultLimit)
	params.Set("songCount", resultLimit)

	result, err := client.callGet("search3", params)
	if err != nil {
		return nil, err
	}
	return result.SubsonicResponse.SearchResult3, nil
}

func (client *NavidromeClient) GetStream(songID string) (io.ReadCloser, string, error) {
	params := url.Values{}
	params.Set("id", songID)

	URL := client.buildURL("stream", params)
	resp, err := client.HttpClient.Get(URL)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, "", fmt.Errorf("server returned error: %s", resp.Status)
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
}
