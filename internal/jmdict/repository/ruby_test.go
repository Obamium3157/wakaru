package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildRubySegments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		surface  string
		reading  string
		expected []RubySegment
	}{
		{
			name:    "hiragana only",
			surface: "ありがとう",
			reading: "ありがとう",
			expected: []RubySegment{
				{Text: "ありがとう"},
			},
		},
		{
			name:    "katakana only",
			surface: "パソコン",
			reading: "パソコン",
			expected: []RubySegment{
				{Text: "パソコン"},
			},
		},
		{
			name:    "single kanji",
			surface: "山",
			reading: "ヤマ",
			expected: []RubySegment{
				{Text: "山", Reading: new("ヤマ")},
			},
		},
		{
			name:    "kanji compound",
			surface: "明日",
			reading: "アシタ",
			expected: []RubySegment{
				{Text: "明日", Reading: new("アシタ")},
			},
		},
		{
			name:    "kanji compound (irregular reading)",
			surface: "大人",
			reading: "オトナ",
			expected: []RubySegment{
				{Text: "大人", Reading: new("オトナ")},
			},
		},
		{
			name:    "mixed kanji-kana; kana in the middle",
			surface: "食べ物",
			reading: "たべもの",
			expected: []RubySegment{
				{Text: "食", Reading: new("た")},
				{Text: "べ"},
				{Text: "物", Reading: new("もの")},
			},
		},
		{
			name:    "mixed kanji-kana; kanji-kana-kanji-kana",
			surface: "取り扱う",
			reading: "トリアツカウ",
			expected: []RubySegment{
				{Text: "取", Reading: new("ト")},
				{Text: "り"},
				{Text: "扱", Reading: new("アツカ")},
				{Text: "う"},
			},
		},
		{
			name:    "mixed kanji-kana",
			surface: "読み上げる",
			reading: "ヨミアゲル",
			expected: []RubySegment{
				{Text: "読", Reading: new("ヨ")},
				{Text: "み"},
				{Text: "上", Reading: new("ア")},
				{Text: "げる"},
			},
		},
		{
			name:    "single kanji with okurigana",
			surface: "読む",
			reading: "ヨム",
			expected: []RubySegment{
				{Text: "読", Reading: new("ヨ")},
				{Text: "む"},
			},
		},
		{
			name:    "kanji three characters",
			surface: "東京都",
			reading: "トウキョウ",
			expected: []RubySegment{
				{Text: "東京都", Reading: new("トウキョウ")},
			},
		},
		{
			name:    "empty surface",
			surface: "",
			reading: "アシタ",
			expected: []RubySegment{
				{Text: ""},
			},
		},
		{
			name:    "empty reading",
			surface: "明日",
			reading: "",
			expected: []RubySegment{
				{Text: "明日"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BuildRubySegments(tt.surface, tt.reading)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuildRubySegmentsDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	surface := "食べ物"
	reading := "たべもの"

	_ = BuildRubySegments(surface, reading)

	assert.Equal(t, "食べ物", surface)
	assert.Equal(t, "たべもの", reading)
}
