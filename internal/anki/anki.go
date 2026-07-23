// Package anki provides a client for the Anki Connect API.
// It supports adding notes to Anki decks over HTTP.
package anki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) AddBasicNote(ctx context.Context, req AddBasicNoteRequest) error {
	payload := formAddBasicNotePayload(req)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}

	httpReq, err := createAddBasicNoteHTTPRequest(
		ctx,
		url.String(),
		body,
	)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("anki request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func formAddBasicNotePayload(reqData AddBasicNoteRequest) map[string]any {
	return map[string]any{
		"action":  "addNote",
		"version": 6,
		"params": map[string]any{
			"note": map[string]any{
				"deckName":  reqData.DeckName,
				"modelName": "Basic",
				"fields": map[string]string{
					"Front": reqData.Front,
					"Back":  reqData.Back,
				},
				"options": map[string]any{
					"allowDuplicate": false,
					"duplicateScope": "deck",
					"duplicateScopeOptions": map[string]any{
						"deckName":       "Default",
						"checkChildren":  false,
						"checkAllModels": false,
					},
				},
				"tags": reqData.Tags,
			},
		},
	}
}

func createAddBasicNoteHTTPRequest(ctx context.Context, url string, body []byte) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	return httpReq, nil
}
