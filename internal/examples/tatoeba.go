// Package examples provides interface for accessing web services with Japanese example sentences
package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	relevanceSO string = "relevance"
	wordsSO            = "words"
	revWordsSO         = "-words"
	createdSO          = "created"
	revCreated         = "-created"
	modifiedSO         = "modified"
	randomSO           = "random"
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

type SearchParameters struct {
	Word         string
	MinWordCount *int
	MaxWordCount *int
	Sort         string
	Limit        *int
}

func (c *Client) Search(ctx context.Context, params SearchParameters) ([]Example, error) {
	url, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	setRawQuery(url, params)

	resp, err := c.getResponse(ctx, url.String())
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
		return nil, fmt.Errorf("tatoeba  error: %s: %s	", resp.Status, body)
	}

	var result tatoebaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	examples := formExamples(result)

	return examples, nil
}

func (c *Client) getResponse(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func setRawQuery(u *url.URL, params SearchParameters) {
	q := u.Query()
	q.Set("lang", langCode)
	q.Set("q", params.Word)
	if params.MinWordCount != nil && params.MaxWordCount == nil {
		q.Set("word_count", fmt.Sprintf("%d-", *params.MinWordCount))
	}
	if params.MinWordCount == nil && params.MinWordCount != nil {
		q.Set("word_count", fmt.Sprintf("-%d", *params.MaxWordCount))
	}
	if params.MinWordCount != nil && params.MaxWordCount != nil {
		minC := *params.MinWordCount
		maxC := *params.MaxWordCount

		if minC == maxC {
			q.Set("word_count", strconv.Itoa(*params.MinWordCount))
		} else {
			q.Set("word_count", fmt.Sprintf("%d-%d", minC, maxC))
		}
	}
	if params.Limit != nil {
		q.Set("limit", strconv.Itoa(*params.Limit))
	}
	q.Set("sort", params.Sort)
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
