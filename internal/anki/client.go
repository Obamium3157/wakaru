package anki

import (
	"fmt"
	"net/http"
	"time"
)

const (
	ankiURL        = "http://localhost"
	defaultTimeout = 12 * time.Second
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(ankiPort int) (*Client, error) {
	if ok := validatePort(ankiPort); !ok {
		return nil, fmt.Errorf("anki port should be in range [1, 65535], got %d", ankiPort)
	}

	baseURL := fmt.Sprintf("%s:%d", ankiURL, ankiPort)

	return &Client{
			baseURL: baseURL,
			client: &http.Client{
				Timeout: defaultTimeout,
			},
		},
		nil
}

func validatePort(port int) bool {
	return 1 <= port && port <= 65535
}
