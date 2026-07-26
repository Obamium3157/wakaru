package examples

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/sentences", r.URL.Path)
		assert.Equal(t, "jpn", r.URL.Query().Get("lang"))
		assert.Equal(t, "試験", r.URL.Query().Get("q"))
		assert.Equal(t, "relevance", r.URL.Query().Get("sort"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{
			"data": [
				{
					"id": 78625,
					"text": "来週は試験だ。",
					"lang": "jpn"
				},
				{
					"id": 13155663,
					"text": "試験を受けます。",
					"lang": "jpn"
				}
			]
		}`))
		require.NoError(t, err)
	}))

	return server
}

func TestClientSearch_Success(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newClientForTesting(server.URL+"/v1/sentences", server.Client())

	got, err := client.Search(context.Background(), SearchParameters{Word: "試験", Sort: "relevance"})

	require.NoError(t, err)
	assert.Equal(t, []Example{
		{
			ID:   78625,
			Text: "来週は試験だ。",
		},
		{
			ID:   13155663,
			Text: "試験を受けます。",
		},
	}, got)
}

func TestClientSearch_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newClientForTesting(server.URL, server.Client())

	got, err := client.Search(context.Background(), SearchParameters{Word: "試験", Sort: "relevance"})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "500")
}

func TestClientSearch_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{invalid json}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := newClientForTesting(server.URL, server.Client())

	got, err := client.Search(context.Background(), SearchParameters{Word: "試験", Sort: "relevance"})

	require.Error(t, err)
	assert.Nil(t, got)
}

func TestClientSearch_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":[]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := newClientForTesting(server.URL, server.Client())

	got, err := client.Search(context.Background(), SearchParameters{Word: "不存在", Sort: "relevance"})

	require.NoError(t, err)
	assert.Empty(t, got)
}
