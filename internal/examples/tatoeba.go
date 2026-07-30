// Package examples provides interface for accessing web services with Japanese example sentences
package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	relevanceSO = "relevance"
	wordsSO     = "words"
	revWordsSO  = "-words"
	createdSO   = "created"
	revCreated  = "-created"
	modifiedSO  = "modified"
	randomSO    = "random"

	amountOfRetries int = 2
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
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type SearchParameters struct {
	Word         string
	MinWordCount *int
	MaxWordCount *int
	Sort         string
	Limit        *int
}

func (c *Client) Search(ctx context.Context, params SearchParameters) ([]Example, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	setRawQuery(u, params)

	var lastErr error

	for attempt := range amountOfRetries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if attempt > 0 {
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		reqStart := time.Now()
		resp, err := c.getResponse(ctx, u.String())
		if err != nil {
			log.Printf(
				"tatoeba attempt %d for %q failed: %v (%v)",
				attempt,
				params.Word,
				err,
				time.Since(reqStart),
			)
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			log.Printf("tatoeba attempt %d for %q: failed to close response body: %v", attempt, params.Word, err)
		}
		if err != nil {
			log.Printf(
				"tatoeba attempt %d for %q: failed to read body: %v (%v)",
				attempt,
				params.Word,
				err,
				time.Since(reqStart),
			)
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("tatoeba error: %s: %s", resp.Status, body)
			log.Printf(
				"tatoeba attempt %d for %q failed: %v (%v)",
				attempt,
				params.Word,
				lastErr,
				time.Since(reqStart),
			)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("tatoeba error: %s: %s", resp.Status, body)
		}

		var result tatoebaResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		return formExamples(result), nil
	}

	return nil, lastErr
}

func (p *SearchParameters) validate() error {
	if p.Word == "" {
		return fmt.Errorf("word should be non-empty string")
	}

	switch p.Sort {
	case relevanceSO, wordsSO, revWordsSO, createdSO, revCreated,
		modifiedSO, randomSO:
		return nil
	default:
		return fmt.Errorf("unknown sort type: %s", p.Sort)
	}
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
	if params.MinWordCount == nil && params.MaxWordCount != nil {
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
