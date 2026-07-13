package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "jpn", r.URL.Query().Get("lang"))
			assert.Equal(t, "", r.URL.Query().Get("q"))
			assert.Equal(t, "relevance", r.URL.Query().Get("sort"))

			w.Header().Set("Content-Type", "application/json")

			_, _ = w.Write([]byte(`
{
	"data":[
		{
			"id":78625,
			"text":"来週は試験だ。",
			"lang":"jpn"
		},
		{
			"id":13155663,
			"text":"試験を受けます。",
			"lang":"jpn"
		}
	]
}
`))
		}),
	)

	defer server.Close()

	old := baseURL
	baseURL = server.URL
	defer func() {
		baseURL = old
	}()

	got, err := Search("")

	require.NoError(t, err)

	assert.Equal(
		t,
		[]Example{
			{
				ID:   78625,
				Text: "来週は試験だ。",
			},
			{
				ID:   13155663,
				Text: "試験を受けます。",
			},
		},
		got,
	)
}
