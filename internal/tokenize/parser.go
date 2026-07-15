package tokenize

import (
	"fmt"
	"strings"

	"github.com/ikawaha/kagome/v2/tokenizer"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindNoun
	KindVerb
	KindAuxVerb
	KindParticle
	KindSymbol
	KindPrefix
	KindAdjective
	KindConjunction
	KindInterjection
	KindAdverb
	KindPrenominal
	KindFiller
	KindOther
)

type RawToken struct {
	Surface  string
	BaseForm string

	POSMajor   Kind
	POSMinor   string
	POSDetail1 string
	POSDetail2 string
}

type DisplayToken struct {
	Surface string
	Lookup  *string
}

func Parse(tokens []tokenizer.Token, lookupSet map[string]bool) ([]DisplayToken, error) {
	p, err := newParser(tokens, lookupSet)
	if err != nil {
		return nil, err
	}

	var result []DisplayToken

	for !p.eof() {
		if tok, ok := p.parseVerbPhrase(); ok {
			result = append(result, tok)
			continue
		}
		if tok, ok := p.parseSuruVerb(); ok {
			result = append(result, tok)
			continue
		}
		if tok, ok := p.parseNaAdjective(); ok {
			result = append(result, tok)
			continue
		}
		if tok, ok := p.parseIAdjective(); ok {
			result = append(result, tok)
			continue
		}
		if tok, ok := p.parseNounPhrase(); ok {
			result = append(result, tok)
			continue
		}
		if tok, ok := p.parseParticlePhrase(); ok {
			result = append(result, tok)
			continue
		}
		if tok, ok := p.parseCompoundNumber(); ok {
			result = append(result, tok)
			continue
		}

		result = append(result, p.parseUnknown())
	}

	return result, nil
}

type parser struct {
	tokens    []RawToken
	pos       int
	lookupSet map[string]bool
}

func (p *parser) eof() bool {
	return p.pos >= len(p.tokens)
}

func (p *parser) peek() *RawToken {
	if p.eof() {
		return nil
	}

	return &p.tokens[p.pos]
}

func (p *parser) next() RawToken {
	t := p.tokens[p.pos]
	p.pos++

	return t
}

func newParser(tokens []tokenizer.Token, lookupSet map[string]bool) (*parser, error) {
	var rawTokens []RawToken
	for _, tok := range tokens {
		rt, err := NewRawToken(tok)
		if err != nil {
			return nil, err
		}
		rawTokens = append(rawTokens, *rt)
	}

	p := &parser{
		tokens:    rawTokens,
		lookupSet: lookupSet,
	}

	return p, nil
}

func NewRawToken(t tokenizer.Token) (*RawToken, error) {
	features := t.Features()
	if len(features) < 4 {
		return nil, fmt.Errorf(
			"expected at least 4 POS features, got %d",
			len(features),
		)
	}

	baseForm := t.Surface
	if bf, ok := t.BaseForm(); ok {
		baseForm = bf
	}

	rt := RawToken{
		Surface:    t.Surface,
		BaseForm:   baseForm,
		POSMinor:   features[1],
		POSDetail1: features[2],
		POSDetail2: features[3],
	}

	if err := rt.initPOSMajor(features[0]); err != nil {
		return nil, err
	}

	return &rt, nil
}

func (t *RawToken) initPOSMajor(kind string) error {
	switch kind {

	case "名詞":
		t.POSMajor = KindNoun
	case "動詞":
		t.POSMajor = KindVerb
	case "形容詞":
		t.POSMajor = KindAdjective
	case "助詞":
		t.POSMajor = KindParticle
	case "助動詞":
		t.POSMajor = KindAuxVerb
	case "記号":
		t.POSMajor = KindSymbol
	case "接頭詞":
		t.POSMajor = KindPrefix
	case "接続詞":
		t.POSMajor = KindConjunction
	case "感動詞":
		t.POSMajor = KindInterjection
	case "副詞":
		t.POSMajor = KindAdverb
	case "連体詞":
		t.POSMajor = KindPrenominal
	case "フィラー":
		t.POSMajor = KindFiller
	case "その他":
		t.POSMajor = KindOther

	default:
		return fmt.Errorf("unknown POS1 %q", kind)
	}

	return nil
}

// parseVerbPhrase covers:
//
//	動詞(自立)
//	動詞(自立) + 助動詞+
//	動詞(自立) + 動詞(接尾)+
//	動詞(自立) + 動詞(非自立)+
//	動詞(自立) + 動詞(接尾)+ + 助動詞+
//	動詞(自立) + 動詞(非自立)+ + 助動詞+
//	動詞(自立) + 接続助詞(て/で) + 動詞(非自立)+
//	動詞(自立) + 接続助詞(て/で) + 形容詞(非自立)
//	動詞(自立) + 形容詞(非自立)
func (p *parser) parseVerbPhrase() (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}

	tok := p.peek()
	if tok.POSMajor != KindVerb || tok.POSMinor != "自立" {
		return DisplayToken{}, false
	}

	start := p.next()
	var surface strings.Builder
	surface.WriteString(start.Surface)

	for !p.eof() {
		next := p.peek()
		consume := false

		switch next.POSMajor {
		case KindVerb:
			consume = next.POSMinor == "非自立" || next.POSMinor == "接尾"
		case KindAuxVerb:
			consume = true
		case KindParticle:
			consume = next.POSMinor == "接続助詞"
		case KindAdjective:
			consume = next.POSMinor == "非自立"
		}

		if !consume {
			break
		}

		surface.WriteString(next.Surface)
		p.next()
	}

	lookup := start.BaseForm
	return DisplayToken{
		Surface: surface.String(),
		Lookup:  &lookup,
	}, true
}

// parseSuruVerb covers:
//
//	名詞(サ変接続) + する
//	名詞(サ変接続) + する + 助動詞+
//	名詞(サ変接続) + する + 動詞(接尾)+
//	名詞(サ変接続) + する + 動詞(非自立)+
//	名詞(サ変接続) + する + 接続助詞(て/で) + 動詞(非自立)+
func (p *parser) parseSuruVerb() (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}

	tok := p.peek()
	if tok.POSMajor != KindNoun || tok.POSMinor != "サ変接続" {
		return DisplayToken{}, false
	}

	startPos := p.pos
	start := p.next()
	var surface strings.Builder
	surface.WriteString(start.Surface)

	if p.eof() || p.peek().BaseForm != "する" {
		p.pos = startPos
		return DisplayToken{}, false
	}
	surface.WriteString(p.next().Surface)

	for !p.eof() {
		next := p.peek()
		consume := false

		switch next.POSMajor {
		case KindVerb:
			consume = next.POSMinor == "非自立" || next.POSMinor == "接尾"
		case KindAuxVerb:
			consume = true
		case KindParticle:
			consume = next.POSMinor == "接続助詞"
		case KindAdjective:
			consume = next.POSMinor == "非自立"
		}

		if !consume {
			break
		}

		surface.WriteString(next.Surface)
		p.next()
	}

	lookup := start.Surface
	return DisplayToken{
		Surface: surface.String(),
		Lookup:  &lookup,
	}, true
}

// parseNaAdjective covers:
//
//	名詞(形容動詞語幹)
//	名詞(形容動詞語幹) + 助動詞
//	名詞(形容動詞語幹) + 助動詞+
func (p *parser) parseNaAdjective() (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}

	tok := p.peek()

	if tok.POSMajor != KindNoun || tok.POSMinor != "形容動詞語幹" {
		return DisplayToken{}, false
	}

	start := p.next()
	var surface strings.Builder
	surface.WriteString(start.Surface)

	for !p.eof() {
		next := p.peek()
		if next.POSMajor != KindAuxVerb {
			break
		}
		surface.WriteString(next.Surface)
		p.next()
	}

	lookup := start.Surface
	return DisplayToken{
		Surface: surface.String(),
		Lookup:  &lookup,
	}, true
}

// parseIAdjective covers:
//
//	形容詞(自立)
//	形容詞(自立) + 助動詞+
func (p *parser) parseIAdjective() (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}

	tok := p.peek()
	if tok.POSMajor != KindAdjective || tok.POSMinor != "自立" {
		return DisplayToken{}, false
	}

	start := p.next()
	var surface strings.Builder
	surface.WriteString(start.Surface)

	for !p.eof() {
		next := p.peek()
		if next.POSMajor != KindAuxVerb {
			break
		}
		surface.WriteString(next.Surface)
		p.next()
	}

	lookup := start.BaseForm
	return DisplayToken{
		Surface: surface.String(),
		Lookup:  &lookup,
	}, true
}

// parseNounPhrase covers:
//
//	接頭詞 + 名詞+
//	名詞+
//	名詞(数)+
//	名詞(数)+ 名詞(接尾,助数詞)
//	名詞(固有名詞)+
func (p *parser) parseNounPhrase() (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}

	tok := p.peek()
	if tok.POSMajor != KindNoun && tok.POSMajor != KindPrefix {
		return DisplayToken{}, false
	}

	startPos := p.pos
	nouns := p.collectNounPhrase()

	if len(p.lookupSet) == 0 {
		p.pos = startPos + 1
		s := tok.Surface
		return DisplayToken{
			Surface: s,
			Lookup:  &s,
		}, true
	}

	if prefix, n := p.findLongestPrefix(nouns); n > 0 {
		p.pos = startPos + n
		return DisplayToken{
			Surface: prefix,
			Lookup:  &prefix,
		}, true
	}

	p.pos = startPos + 1
	s := tok.Surface
	return DisplayToken{
		Surface: s,
		Lookup:  &s,
	}, true
}

func (p *parser) collectNounPhrase() []RawToken {
	var nouns []RawToken
	for i := p.pos; i < len(p.tokens); i++ {
		if p.tokens[i].POSMajor != KindNoun && p.tokens[i].POSMajor != KindPrefix {
			break
		}
		nouns = append(nouns, p.tokens[i])
	}
	return nouns
}

func (p *parser) findLongestPrefix(nouns []RawToken) (string, int) {
	for i := len(nouns); i >= 1; i-- {
		var prefix strings.Builder
		for j := 0; j < i; j++ {
			prefix.WriteString(nouns[j].Surface)
		}
		prefixStr := prefix.String()
		if p.lookupSet[prefixStr] {
			return prefixStr, i
		}
	}
	return "", 0
}

// parseParticlePhrase covers:
//
//	助詞+
func (p *parser) parseParticlePhrase() (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}

	tok := p.peek()
	if tok.POSMajor != KindParticle {
		return DisplayToken{}, false
	}

	var surface strings.Builder

	for !p.eof() {
		next := p.peek()
		if next.POSMajor != KindParticle {
			break
		}
		surface.WriteString(next.Surface)
		p.next()
	}

	result := surface.String()
	return DisplayToken{
		Surface: result,
		Lookup:  nil,
	}, true
}

// parseCompoundNumber covers:
//
//	名詞(数)+
//	名詞(数)+ 名詞(接尾,助数詞)
func (p *parser) parseCompoundNumber() (DisplayToken, bool) {
	if p.eof() {
		return DisplayToken{}, false
	}

	tok := p.peek()
	if tok.POSMajor != KindNoun || tok.POSMinor != "数" {
		return DisplayToken{}, false
	}

	var surface strings.Builder

	for !p.eof() {
		next := p.peek()
		if next.POSMajor != KindNoun {
			break
		}
		if next.POSMinor == "数" || (next.POSMinor == "接尾" && next.POSDetail1 == "助数詞") {
			surface.WriteString(next.Surface)
			p.next()
		} else {
			break
		}
	}

	result := surface.String()
	return DisplayToken{
		Surface: result,
		Lookup:  &result,
	}, true
}

// parseUnknown consumes a single token that was not recognized by any
// specialized parser.
func (p *parser) parseUnknown() DisplayToken {
	tok := p.next()
	s := tok.Surface
	return DisplayToken{
		Surface: s,
		Lookup:  nil,
	}
}
