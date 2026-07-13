// Package examples provides interface for accessing web services with Japanese example sentences
package examples

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	baseURL  = "https://api.tatoeba.org/v1/sentences"
	langCode = "jpn"
)

type sentence struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Lang string `json:"lang"`
}

type tatoebaResponse struct {
	Data []sentence `json:"data"`
}

type Example struct {
	ID   int
	Text string
}

func Search(word string) ([]Example, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	setRawQuery(u, word)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tatoeba error: %s: %s", resp.Status, body)
	}

	var result tatoebaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	examples := formExamples(result)

	return examples, nil
}

func setRawQuery(u *url.URL, word string) {
	q := u.Query()
	q.Set("lang", langCode)
	q.Set("q", word)
	q.Set("sort", "relevance")
	u.RawQuery = q.Encode()
}

func formExamples(resp tatoebaResponse) []Example {
	examples := make([]Example, 0, len(resp.Data))

	for _, item := range resp.Data {
		examples = append(examples, Example{
			ID:   item.ID,
			Text: item.Text,
		})
	}

	return examples
}
