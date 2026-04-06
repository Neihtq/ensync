// Package navidrome implements interfaces to communicate with the Subsonic API
// Subsonic API refs: https://www.subsonic.org/pages/api.jsp
package navidrome

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type ResponseWrapper struct {
	SubsonicResponse SubsonicResponse `json:"subsonic-response"`
}

type SubsonicResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Error   *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type NavidromeClient struct {
	BaseURL    string
	ApiVersion string
	ClientName string

	HttpClient *http.Client
}

func NewNavidromeClient() *NavidromeClient {
	serverURL := os.Getenv("NAVIDROME_URL")
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
	var result ResponseWrapper
	err := client.callGet("ping", nil, &result)
	if err != nil {
		return err
	}

	if result.SubsonicResponse.Status == "failed" {
		if result.SubsonicResponse.Error != nil {
			return fmt.Errorf("API error %d: %s", result.SubsonicResponse.Error.Code, result.SubsonicResponse.Error.Message)
		}
		return fmt.Errorf("API failed with an unknown error")
	}

	return nil
}

func (client *NavidromeClient) buildURL(endpoint string, extraParams url.Values) string {
	base, _ := url.Parse(fmt.Sprintf("%s/rest/%s", client.BaseURL, endpoint))

	params := client.buildauthParms()
	for key, values := range extraParams {
		for _, val := range values {
			params.Add(key, val)
		}
	}

	base.RawQuery = params.Encode()
	return base.String()
}

func (client *NavidromeClient) callGet(endpoint string, params url.Values, target interface{}) error {
	URL := client.buildURL(endpoint, params)

	resp, err := client.HttpClient.Get(URL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}

func getCredentials() (string, string) {
	user := os.Getenv("NAVIDROME_USER")
	password := os.Getenv("NAVIDROME_PASSWORD")

	return user, password
}
