package tokenize

import (
	"testing"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	p, err := NewParser(tokens, nil)
	assert.NoError(t, err)

	return p
}

// func createTestParserWithLookup(t *testing.T, input string, lookupSet map[string]bool) *parser {
// 	t.Helper()
//
// 	tokens := extractKagomeTokens(t, input)
// 	p, err := NewParser(tokens, lookupSet)
// 	assert.NoError(t, err)
//
// 	return p
// }

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
	t.Parallel()

	s := "食べます"
	p := createTestParser(t, s)

	token, ok := p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "食べる", *token.Lookup)
}

func TestParseVerbPhraseMultipleSuffixes(t *testing.T) {
	t.Parallel()

	s := "食べさせられました"
	p := createTestParser(t, s)

	token, ok := p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "食べる", *token.Lookup)
}

func TestParseVerbPhraseTeForm(t *testing.T) {
	t.Parallel()

	s := "食べている"
	p := createTestParser(t, s)

	token, ok := p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "食べる", *token.Lookup)

	s = "食べておく"
	p = createTestParser(t, s)

	token, ok = p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "食べる", *token.Lookup)

	s = "食べてしまう"
	p = createTestParser(t, s)

	token, ok = p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "食べる", *token.Lookup)

	s = "読んでしまう"
	p = createTestParser(t, s)

	token, ok = p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "読む", *token.Lookup)
}

func TestParseVerbPhraseYasuiNikui(t *testing.T) {
	t.Parallel()

	s := "読みやすい"
	p := createTestParser(t, s)

	token, ok := p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "読む", *token.Lookup)

	s = "読みにくい"
	p = createTestParser(t, s)

	token, ok = p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "読む", *token.Lookup)
}

func TestParseVerbPhraseNakuNaru(t *testing.T) {
	t.Parallel()

	s := "食べなくなる"
	p := createTestParser(t, s)

	token, ok := p.parseVerbPhrase()
	assert.True(t, ok)
	assert.Equal(t, "食べなく", token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "食べる", *token.Lookup)
}

func TestParseSuruVerb(t *testing.T) {
	t.Parallel()

	s := "勉強する"
	p := createTestParser(t, s)

	token, ok := p.parseSuruVerb()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "勉強", *token.Lookup)

	s = "勉強させられている"
	p = createTestParser(t, s)

	token, ok = p.parseSuruVerb()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "勉強", *token.Lookup)
}

func TestParsePlainNaAdjective(t *testing.T) {
	t.Parallel()

	s := "静かなデン"
	p := createTestParser(t, s)

	token, ok := p.parseNaAdjective()
	assert.True(t, ok)
	assert.Equal(t, "静かな", token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "静か", *token.Lookup)

	_, ok = p.parseNaAdjective()
	assert.False(t, ok)
}

func TestParseNaAdjectiveWithSuffixes(t *testing.T) {
	t.Parallel()

	s := "便利でした"
	p := createTestParser(t, s)

	token, ok := p.parseNaAdjective()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "便利", *token.Lookup)
}

func TestParseNaAdjective_Negative(t *testing.T) {
	t.Parallel()

	s := "名前"
	p := createTestParser(t, s)

	token, ok := p.parseNaAdjective()
	assert.False(t, ok)
	assert.Equal(t, DisplayToken{}, token)
}

func TestParseIAdjective(t *testing.T) {
	t.Parallel()

	s := "低い声"
	p := createTestParser(t, s)

	token, ok := p.parseIAdjective()
	assert.True(t, ok)
	assert.Equal(t, "低い", token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "低い", *token.Lookup)

	s = "安くない"
	p = createTestParser(t, s)

	token, ok = p.parseIAdjective()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "安い", *token.Lookup)

	s = "高くなかった"
	p = createTestParser(t, s)

	token, ok = p.parseIAdjective()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, "高い", *token.Lookup)
}

func TestParseIAdjectiveNaru(t *testing.T) {
	t.Parallel()

	s := "安くならない"
	p := createTestParser(t, s)

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
	t.Parallel()

	s := "高校生"
	p := createTestParser(t, s)

	token, ok := p.parseNounPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, s, *token.Lookup)

	s = "ご家族"
	p = createTestParser(t, s)
	token, ok = p.parseNounPhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, s, *token.Lookup)
}

func TestParseParticlePhrase(t *testing.T) {
	t.Parallel()

	s := "は"
	p := createTestParser(t, s)

	token, ok := p.parseParticlePhrase()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	assert.Nil(t, token.Lookup)
}

func TestParseCompoundNumber(t *testing.T) {
	t.Parallel()

	s := "二万三千六百二十五"
	p := createTestParser(t, s)

	token, ok := p.parseCompoundNumber()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, s, *token.Lookup)
}

func TestParseCompoundNumberCounters(t *testing.T) {
	t.Parallel()

	s := "五本"
	p := createTestParser(t, s)

	token, ok := p.parseCompoundNumber()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, s, *token.Lookup)

	s = "三日"
	p = createTestParser(t, s)

	token, ok = p.parseCompoundNumber()
	assert.True(t, ok)
	assert.Equal(t, s, token.Surface)
	require.NotNil(t, token.Lookup)
	assert.Equal(t, s, *token.Lookup)
}

func TestParseUnknown(t *testing.T) {
	t.Parallel()

	s := "、"
	p := createTestParser(t, s)

	token := p.parseUnknown()
	assert.Equal(t, s, token.Surface)
	assert.Nil(t, token.Lookup)
}
