package navidrome

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

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

func (client *NavidromeClient) callGet(endpoint string, params url.Values) (*ResponseWrapper, error) {
	var result ResponseWrapper
	URL := client.buildURL(endpoint, params)

	resp, err := client.HttpClient.Get(URL)
	if err != nil {
		return &result, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &result, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	if result.SubsonicResponse.Status == "failed" {
		if result.SubsonicResponse.Error != nil {
			return &result, fmt.Errorf("API error %d: %s", result.SubsonicResponse.Error.Code, result.SubsonicResponse.Error.Message)
		}
		return &result, fmt.Errorf("API failed with an unknown error")
	}

	return &result, nil
}

func getCredentials() (string, string) {
	user := os.Getenv("NAVIDROME_USER")
	password := os.Getenv("NAVIDROME_PASSWORD")

	if user == "" || password == "" {
		panic("Navidrome credentials not found!")
	}

	return user, password
}
