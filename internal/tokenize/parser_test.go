package tokenize

import (
	"testing"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parseCase struct {
	name      string
	input     string
	lookupSet map[string]bool
	want      []tokenExpect
}

type tokenExpect struct {
	Surface  string
	Lookup   *string
	POSMajor Kind
}

func expect(surface string, lookup string, posMajor Kind) tokenExpect {
	return tokenExpect{
		surface,
		&lookup,
		posMajor,
	}
}

func runParseTests(t *testing.T, parse func(*parser) (DisplayToken, bool), cases []parseCase) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := createTestParserWithLookup(t, tc.input, tc.lookupSet)

			if tc.want == nil {
				_, ok := parse(p)
				assert.False(t, ok)
				return
			}

			for i, w := range tc.want {
				tok, ok := parse(p)
				require.True(t, ok, "case %d: expected parse to succeed", i)
				assert.Equal(t, w.Surface, tok.Surface, "case %d: surface mismatch", i)

				if w.Lookup != nil {
					require.NotNil(t, tok.Lookup, "case %d: expected lookup %q, got nil", i, *w.Lookup)
					assert.Equal(t, *w.Lookup, *tok.Lookup, "case %d: lookup mismatch", i)
				} else {
					assert.Nil(t, tok.Lookup, "case %d: expected nil lookup", i)
				}
			}

			_, ok := parse(p)
			assert.False(t, ok, "expected parser to be exhausted, but more tokens remain")
		})
	}
}

func extractKagomeTokens(t *testing.T, input string) []tokenizer.Token {
	t.Helper()

	tr, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	assert.NoError(t, err)

	tokens := tr.Tokenize(input)

	return tokens
}

func createTestParser(t *testing.T, input string) *parser {
	t.Helper()

	tokens := extractKagomeTokens(t, input)
	p, err := newParser(tokens, nil)
	assert.NoError(t, err)

	return p
}

func createTestParserWithLookup(t *testing.T, input string, lookupSet map[string]bool) *parser {
	t.Helper()

	tokens := extractKagomeTokens(t, input)
	p, err := newParser(tokens, lookupSet)
	assert.NoError(t, err)

	return p
}

func TestFindLongestPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		surfaces []string
		lookup   map[string]bool
		wantStr  string
		wantN    int
	}{
		{
			name:     "empty nouns slice",
			surfaces: []string{},
			lookup:   map[string]bool{"A": true},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "nil nouns slice",
			surfaces: nil,
			lookup:   map[string]bool{"A": true},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "single noun found",
			surfaces: []string{"A"},
			lookup:   map[string]bool{"A": true},
			wantStr:  "A",
			wantN:    1,
		},
		{
			name:     "single noun not found",
			surfaces: []string{"A"},
			lookup:   map[string]bool{},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "two nouns full match",
			surfaces: []string{"A", "B"},
			lookup:   map[string]bool{"AB": true},
			wantStr:  "AB",
			wantN:    2,
		},
		{
			name:     "two nouns only first matches",
			surfaces: []string{"A", "B"},
			lookup:   map[string]bool{"A": true},
			wantStr:  "A",
			wantN:    1,
		},
		{
			name:     "two nouns none match",
			surfaces: []string{"A", "B"},
			lookup:   map[string]bool{},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "two nouns first not found but full found",
			surfaces: []string{"A", "B"},
			lookup:   map[string]bool{"AB": true, "B": true},
			wantStr:  "AB",
			wantN:    2,
		},
		{
			name:     "three nouns full match",
			surfaces: []string{"A", "B", "C"},
			lookup:   map[string]bool{"ABC": true},
			wantStr:  "ABC",
			wantN:    3,
		},
		{
			name:     "three nouns prefix of two matches",
			surfaces: []string{"A", "B", "C"},
			lookup:   map[string]bool{"AB": true},
			wantStr:  "AB",
			wantN:    2,
		},
		{
			name:     "three nouns only first matches",
			surfaces: []string{"A", "B", "C"},
			lookup:   map[string]bool{"A": true},
			wantStr:  "A",
			wantN:    1,
		},
		{
			name:     "three nouns none match",
			surfaces: []string{"A", "B", "C"},
			lookup:   map[string]bool{},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "longest wins over shorter",
			surfaces: []string{"A", "B", "C"},
			lookup:   map[string]bool{"ABC": true, "AB": true, "A": true},
			wantStr:  "ABC",
			wantN:    3,
		},
		{
			name:     "no mid-match only shortest",
			surfaces: []string{"A", "B", "C"},
			lookup:   map[string]bool{"A": true, "C": true},
			wantStr:  "A",
			wantN:    1,
		},
		{
			name:     "substring but not exact",
			surfaces: []string{"A", "B"},
			lookup:   map[string]bool{"ABX": true},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "nil lookupSet",
			surfaces: []string{"A", "B"},
			lookup:   nil,
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "empty lookupSet",
			surfaces: []string{"A", "B"},
			lookup:   map[string]bool{},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "japanese two nouns match",
			surfaces: []string{"毎", "朝"},
			lookup:   map[string]bool{"毎朝": true},
			wantStr:  "毎朝",
			wantN:    2,
		},
		{
			name:     "japanese three nouns split by dictionary",
			surfaces: []string{"毎", "朝", "ご飯"},
			lookup:   map[string]bool{"毎朝": true, "朝ご飯": true},
			wantStr:  "毎朝",
			wantN:    2,
		},
		{
			name:     "japanese three nouns full match",
			surfaces: []string{"毎", "朝", "ご飯"},
			lookup:   map[string]bool{"毎朝ご飯": true},
			wantStr:  "毎朝ご飯",
			wantN:    3,
		},
		{
			name:     "single hiragana char",
			surfaces: []string{"あ"},
			lookup:   map[string]bool{"あ": true},
			wantStr:  "あ",
			wantN:    1,
		},
		{
			name:     "many nouns longest chain",
			surfaces: []string{"A", "B", "C", "D", "E"},
			lookup:   map[string]bool{"ABCDE": true, "ABCD": true, "AB": true},
			wantStr:  "ABCDE",
			wantN:    5,
		},
		{
			name:     "many nouns only second prefix matches",
			surfaces: []string{"A", "B", "C", "D", "E"},
			lookup:   map[string]bool{"AB": true},
			wantStr:  "AB",
			wantN:    2,
		},
		{
			name:     "many nouns none match",
			surfaces: []string{"A", "B", "C", "D", "E"},
			lookup:   map[string]bool{},
			wantStr:  "",
			wantN:    0,
		},
		{
			name:     "japanese kanji nouns full match",
			surfaces: []string{"日本", "語", "能力", "試験"},
			lookup:   map[string]bool{"日本語能力試験": true},
			wantStr:  "日本語能力試験",
			wantN:    4,
		},
		{
			name:     "japanese kanji nouns partial match",
			surfaces: []string{"日本", "語", "能力", "試験"},
			lookup:   map[string]bool{"日本語": true, "能力試験": true},
			wantStr:  "日本語",
			wantN:    2,
		},
		{
			name:     "single noun with empty surface",
			surfaces: []string{""},
			lookup:   map[string]bool{"": true},
			wantStr:  "",
			wantN:    1,
		},
		{
			name:     "mixed length surfaces full match",
			surfaces: []string{"AB", "CDEF", "GH"},
			lookup:   map[string]bool{"ABCDEF": true, "ABCDEFGH": true},
			wantStr:  "ABCDEFGH",
			wantN:    3,
		},
		{
			name:     "mixed length surfaces prefix only",
			surfaces: []string{"AB", "CDEF", "GH"},
			lookup:   map[string]bool{"ABCDEF": true},
			wantStr:  "ABCDEF",
			wantN:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nouns := make([]RawToken, len(tt.surfaces))
			for i, s := range tt.surfaces {
				nouns[i] = RawToken{Surface: s}
			}

			p := &parser{lookupSet: tt.lookup}
			gotStr, gotN := p.findLongestPrefix(nouns)

			assert.Equal(t, tt.wantStr, gotStr, "prefix mismatch")
			assert.Equal(t, tt.wantN, gotN, "noun count mismatch")
		})
	}
}

func TestExtractKagomeTokens(t *testing.T) {
	t.Parallel()

	s := "私はピーターです。"
	tokens := extractKagomeTokens(t, s)
	assert.Equal(t, 5, len(tokens))
	assert.Equal(t, "私", tokens[0].Surface)
	assert.Equal(t, "は", tokens[1].Surface)
	assert.Equal(t, "ピーター", tokens[2].Surface)
	assert.Equal(t, "です", tokens[3].Surface)
	assert.Equal(t, "。", tokens[4].Surface)
}

func TestParseVerbPhrase(t *testing.T) {
	runParseTests(t, (*parser).parseVerbPhrase, []parseCase{
		{"basic", "食べます", nil, []tokenExpect{
			expect("食べます", "食べる", KindVerb),
		}},
		{"multiple suffixes", "食べさせられました", nil, []tokenExpect{
			expect("食べさせられました", "食べる", KindVerb),
		}},
		{"te form", "食べている", nil, []tokenExpect{
			expect("食べている", "食べる", KindVerb),
		}},
		{"te form ておく", "食べておく", nil, []tokenExpect{
			expect("食べておく", "食べる", KindVerb),
		}},
		{"te form てしまう", "食べてしまう", nil, []tokenExpect{
			expect("食べてしまう", "食べる", KindVerb),
		}},
		{"te form another verb", "読んでしまう", nil, []tokenExpect{
			expect("読んでしまう", "読む", KindVerb),
		}},
		{"yasui suffix", "読みやすい", nil, []tokenExpect{
			expect("読みやすい", "読む", KindVerb),
		}},
		{"nikui suffix", "読みにくい", nil, []tokenExpect{
			expect("読みにくい", "読む", KindVerb),
		}},
		{"nai verb that is an adjective", "詰まらない", map[string]bool{"詰まらない": true}, []tokenExpect{
			expect("詰まらない", "詰まらない", KindVerb),
		}},
		{"nai verb that is not an adjective", "走れなくなった", nil, []tokenExpect{
			expect("走れなく", "走れる", KindVerb),
			expect("なった", "なる", KindVerb),
		}},
		{"go-dan potential form", "話せる", map[string]bool{"話す": true}, []tokenExpect{
			expect("話せる", "話す", KindVerb),
		}},
		{"pseudo-go-dan potential verb", "食べる", map[string]bool{"食べる": true}, []tokenExpect{
			expect("食べる", "食べる", KindVerb),
		}},
	})
}

func TestGetPotentialGoDanInfinitive(t *testing.T) {
	expected := []struct {
		Surface  string
		BaseForm string
	}{
		{
			"話せる",
			"話す",
		},
		{
			"聞ける",
			"聞く",
		},
		{
			"泳げる",
			"泳ぐ",
		},
		{
			"遊べる",
			"遊ぶ",
		},
		{
			"待てる",
			"待つ",
		},
		{
			"飲める",
			"飲む",
		},
		{
			"買える",
			"買う",
		},
		{
			"死ねる",
			"死ぬ",
		},
	}

	for _, e := range expected {
		bf := getPotentialGoDanInfinitive(e.Surface)
		assert.Equal(t, e.BaseForm, bf)
	}
}

func TestParseSuruVerb(t *testing.T) {
	runParseTests(t, (*parser).parseSuruVerb, []parseCase{
		{"basic", "勉強する", nil, []tokenExpect{
			expect("勉強する", "勉強", KindVerb),
		}},
		{"with causative passive", "勉強させられている", nil, []tokenExpect{
			expect("勉強させられている", "勉強", KindVerb),
		}},
		{"negative: standalone sahen noun", "発表", nil, nil},
	})
}

func TestParseNaAdjective(t *testing.T) {
	runParseTests(t, (*parser).parseNaAdjective, []parseCase{
		{"plain", "静かなデン", nil, []tokenExpect{
			expect("静かな", "静か", KindAdjective),
		}},
		{"with deshita", "便利でした", nil, []tokenExpect{
			expect("便利でした", "便利", KindAdjective),
		}},
		{"negative: noun is not na adjective", "名前", nil, nil},
	})
}

func TestParseIAdjective(t *testing.T) {
	runParseTests(t, (*parser).parseIAdjective, []parseCase{
		{"followed by noun", "低い声", nil, []tokenExpect{
			expect("低い", "低い", KindAdjective),
		}},
		{"negative", "安くない", nil, []tokenExpect{
			expect("安くない", "安い", KindAdjective),
		}},
		{"past negative", "高くなかった", nil, []tokenExpect{
			expect("高くなかった", "高い", KindAdjective),
		}},
	})
}

func TestParseIAdjectiveNaru(t *testing.T) {
	p := createTestParser(t, "安くならない")

	token, ok := p.parseIAdjective()
	assert.True(t, ok)
	assert.Equal(t, "安く", token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "安い", *token.Lookup)

	token, ok = p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, "ならない", token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "なる", *token.Lookup)
}

func TestParseNounPhrase(t *testing.T) {
	runParseTests(t, (*parser).parseNounPhrase, []parseCase{
		{"single noun", "高校生", nil, []tokenExpect{
			expect("高校生", "高校生", KindNoun),
		}},
		{"prefix plus noun", "ご家族", map[string]bool{"ご家族": true}, []tokenExpect{
			expect("ご家族", "ご家族", KindNoun),
		}},
		{"split by dictionary", "毎朝朝ご飯", map[string]bool{"毎朝": true, "朝ご飯": true}, []tokenExpect{
			expect("毎朝", "毎朝", KindNoun),
			expect("朝ご飯", "朝ご飯", KindNoun),
		}},
		{"JLPT (do not split by dictionary)", "日本語能力試験", nil, []tokenExpect{
			expect("日本語", "日本語", KindNoun),
			expect("能力", "能力", KindNoun),
			expect("試験", "試験", KindNoun),
		}},
		{"JLPT (split by dictionary)", "日本語能力試験", map[string]bool{"日本語能力試験": true}, []tokenExpect{
			expect("日本語能力試験", "日本語能力試験", KindNoun),
		}},
	})
}

func TestParseParticlePhrase(t *testing.T) {
	runParseTests(t, (*parser).parseParticlePhrase, []parseCase{
		{"single particle", "は", nil, []tokenExpect{
			expect("は", "は", KindParticle),
		}},
	})
}

func TestParseCompoundNumber(t *testing.T) {
	runParseTests(t, (*parser).parseCompoundNumber, []parseCase{
		{"large number", "二万三千六百二十五", nil, []tokenExpect{
			expect("二万三千六百二十五", "二万三千六百二十五", KindNoun),
		}},
		{"number plus counter hon", "五本", nil, []tokenExpect{
			expect("五本", "五本", KindNoun),
		}},
		{"number plus counter hi", "三日", nil, []tokenExpect{
			expect("三日", "三日", KindNoun),
		}},
		{"number plus counter nin", "三人", nil, []tokenExpect{
			expect("三人", "三人", KindNoun),
		}},
		{"number plus counter mai", "二枚", nil, []tokenExpect{
			expect("二枚", "二枚", KindNoun),
		}},
		{"single digit", "百", nil, []tokenExpect{
			expect("百", "百", KindNoun),
		}},
		{"negative: counter without number", "枚", nil, nil},
		{"negative: noun is not number", "高校生", nil, nil},
	})
}

func parseUnknownAdapter(p *parser) (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}
	return p.parseUnknown(), true
}

func TestParseUnknows_Words(t *testing.T) {
	runParseTests(t, parseUnknownAdapter, []parseCase{
		{"noun", "虫", map[string]bool{"虫": true}, []tokenExpect{
			expect("虫", "虫", KindNoun),
		}},
		{"verb", "走る", map[string]bool{"走る": true}, []tokenExpect{
			expect("走る", "走る", KindVerb),
		}},
		{"auxverb", "です", map[string]bool{"です": true}, []tokenExpect{
			expect("です", "です", KindAuxVerb),
		}},
		{"particle", "は", map[string]bool{"は": true}, []tokenExpect{
			expect("は", "は", KindParticle),
		}},
		{"adjective", "美味しい", map[string]bool{"美味しい": true}, []tokenExpect{
			expect("美味しい", "美味しい", KindAdjective),
		}},
		{"conjunction", "しかし", map[string]bool{"しかし": true}, []tokenExpect{
			expect("しかし", "しかし", KindConjunction),
		}},
		{"interjection", "あれ", map[string]bool{"あれ": true}, []tokenExpect{
			expect("あれ", "あれ", KindInterjection),
		}},
		{"adverb", "とても", map[string]bool{"とても": true}, []tokenExpect{
			expect("とても", "とても", KindAdverb),
		}},
		{"prenominal", "この", map[string]bool{"この": true}, []tokenExpect{
			expect("この", "この", KindPrenominal),
		}},
	})
}

func TestParseUnknown_Symbols(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"comma", "、"},
		{"period", "。"},
		{"exclamation", "！"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := createTestParser(t, tt.input)
			token := p.parseUnknown()
			assert.Equal(t, "", token.Surface)
			assert.Nil(t, token.Lookup)
		})
	}
}
