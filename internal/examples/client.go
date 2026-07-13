package examples

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.tatoeba.org/v1/sentences"
	langCode       = "jpn"
	defaultTimeout = 10 * time.Second
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func newClientForTesting(baseURL string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	return &Client{
		baseURL,
		client,
	}
}
