package examples

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetRawQuery(t *testing.T) {
	u, err := url.Parse("https://example.com")
	assert.NoError(t, err)

	setRawQuery(u, "試験")

	assert.Equal(
		t,
		"lang=jpn&q=%E8%A9%A6%E9%A8%93&sort=relevance",
		u.RawQuery,
	)
}

func TestFormExamples(t *testing.T) {
	resp := tatoebaResponse{
		Data: []sentence{
			{
				ID:   1,
				Text: "試験があります。",
				Lang: "jpn",
			},
			{
				ID:   2,
				Text: "来週は試験だ。",
				Lang: "jpn",
			},
		},
	}

	got := formExamples(resp)

	assert.Equal(
		t,
		[]Example{
			{ID: 1, Text: "試験があります。"},
			{ID: 2, Text: "来週は試験だ。"},
		},
		got,
	)
}
