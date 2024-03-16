// sky-accounts/pkg/clientlib/accountslib/client.go
package accountslib

import (
	"net/http"
	"time"
)

// Client represents an HTTP client that can be used to send requests to the skills server.
type Client struct {
	BaseURL    string
	HttpClient *http.Client
	Token      string
	ApiKey     string
}

func NewClient(baseURL string, token string, apiKey string, httpClient ...*http.Client) *Client {
	var client *http.Client
	if len(httpClient) > 0 {
		client = httpClient[0]
	} else {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	return &Client{
		BaseURL:    baseURL,
		HttpClient: client,
		Token:      token,
		ApiKey:     apiKey,
	}
}
