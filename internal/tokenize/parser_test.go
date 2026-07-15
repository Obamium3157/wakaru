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
	Surface string
	Lookup  *string
}

func expect(surface string, lookup string) tokenExpect {
	return tokenExpect{
		surface,
		&lookup,
	}
}

func expectNoLookup(surface string) tokenExpect {
	return tokenExpect{
		Surface: surface,
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
			expect("食べます", "食べる"),
		}},
		{"multiple suffixes", "食べさせられました", nil, []tokenExpect{
			expect("食べさせられました", "食べる"),
		}},
		{"te form", "食べている", nil, []tokenExpect{
			expect("食べている", "食べる"),
		}},
		{"te form ておく", "食べておく", nil, []tokenExpect{
			expect("食べておく", "食べる"),
		}},
		{"te form てしまう", "食べてしまう", nil, []tokenExpect{
			expect("食べてしまう", "食べる"),
		}},
		{"te form another verb", "読んでしまう", nil, []tokenExpect{
			expect("読んでしまう", "読む"),
		}},
		{"yasui suffix", "読みやすい", nil, []tokenExpect{
			expect("読みやすい", "読む"),
		}},
		{"nikui suffix", "読みにくい", nil, []tokenExpect{
			expect("読みにくい", "読む"),
		}},
	})
}

func TestParseSuruVerb(t *testing.T) {
	runParseTests(t, (*parser).parseSuruVerb, []parseCase{
		{"basic", "勉強する", nil, []tokenExpect{
			expect("勉強する", "勉強"),
		}},
		{"with causative passive", "勉強させられている", nil, []tokenExpect{
			expect("勉強させられている", "勉強"),
		}},
		{"negative: standalone sahen noun", "発表", nil, nil},
	})
}

func TestParseNaAdjective(t *testing.T) {
	runParseTests(t, (*parser).parseNaAdjective, []parseCase{
		{"plain", "静かなデン", nil, []tokenExpect{
			expect("静かな", "静か"),
		}},
		{"with deshita", "便利でした", nil, []tokenExpect{
			expect("便利でした", "便利"),
		}},
		{"negative: noun is not na adjective", "名前", nil, nil},
	})
}

func TestParseIAdjective(t *testing.T) {
	runParseTests(t, (*parser).parseIAdjective, []parseCase{
		{"followed by noun", "低い声", nil, []tokenExpect{
			expect("低い", "低い"),
		}},
		{"negative", "安くない", nil, []tokenExpect{
			expect("安くない", "安い"),
		}},
		{"past negative", "高くなかった", nil, []tokenExpect{
			expect("高くなかった", "高い"),
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
			expect("高校生", "高校生"),
		}},
		{"prefix plus noun", "ご家族", map[string]bool{"ご家族": true}, []tokenExpect{
			expect("ご家族", "ご家族"),
		}},
		{"split by dictionary", "毎朝朝ご飯", map[string]bool{"毎朝": true, "朝ご飯": true}, []tokenExpect{
			expect("毎朝", "毎朝"),
			expect("朝ご飯", "朝ご飯"),
		}},
		{"JLPT (do not split by dictionary)", "日本語能力試験", nil, []tokenExpect{
			expect("日本語", "日本語"),
			expect("能力", "能力"),
			expect("試験", "試験"),
		}},
		{"JLPT (split by dictionary)", "日本語能力試験", map[string]bool{"日本語能力試験": true}, []tokenExpect{
			expect("日本語能力試験", "日本語能力試験"),
		}},
	})
}

func TestParseParticlePhrase(t *testing.T) {
	runParseTests(t, (*parser).parseParticlePhrase, []parseCase{
		{"single particle", "は", nil, []tokenExpect{
			expectNoLookup("は"),
		}},
	})
}

func TestParseCompoundNumber(t *testing.T) {
	runParseTests(t, (*parser).parseCompoundNumber, []parseCase{
		{"large number", "二万三千六百二十五", nil, []tokenExpect{
			expect("二万三千六百二十五", "二万三千六百二十五"),
		}},
		{"number plus counter hon", "五本", nil, []tokenExpect{
			expect("五本", "五本"),
		}},
		{"number plus counter hi", "三日", nil, []tokenExpect{
			expect("三日", "三日"),
		}},
		{"number plus counter nin", "三人", nil, []tokenExpect{
			expect("三人", "三人"),
		}},
		{"number plus counter mai", "二枚", nil, []tokenExpect{
			expect("二枚", "二枚"),
		}},
		{"single digit", "百", nil, []tokenExpect{
			expect("百", "百"),
		}},
		{"negative: counter without number", "枚", nil, nil},
		{"negative: noun is not number", "高校生", nil, nil},
	})
}

func TestParseUnknown(t *testing.T) {
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
			assert.Equal(t, tt.input, token.Surface)
			assert.Nil(t, token.Lookup)
		})
	}
}
