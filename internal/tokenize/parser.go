//go:generate stringer -type=Kind
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

var posMap = map[string]Kind{
	"名詞":   KindNoun,
	"動詞":   KindVerb,
	"形容詞":  KindAdjective,
	"助詞":   KindParticle,
	"助動詞":  KindAuxVerb,
	"記号":   KindSymbol,
	"接頭詞":  KindPrefix,
	"接続詞":  KindConjunction,
	"感動詞":  KindInterjection,
	"副詞":   KindAdverb,
	"連体詞":  KindPrenominal,
	"フィラー": KindFiller,
	"その他":  KindOther,
}

type RawToken struct {
	Surface  string
	BaseForm string

	POSMajor   Kind
	POSMinor   string
	POSDetail1 string
	POSDetail2 string
}

type DisplayToken struct {
	Surface  string
	Lookup   *string
	POSMajor Kind
}

func Parse(tokens []tokenizer.Token, lookupSet map[string]bool) ([]DisplayToken, error) {
	p, err := newParser(tokens, lookupSet)
	if err != nil {
		return nil, err
	}

	var result []DisplayToken
	methods := p.getMethods()

	for !p.eof() {
		matched := false

		for _, parse := range methods {
			if tok, ok := parse(); ok {
				result = append(result, tok)
				matched = true
				break
			}
		}

		if !matched {
			result = append(result, p.parseUnknown())
		}
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

type parserMethod func() (DisplayToken, bool)

func (p *parser) getMethods() []parserMethod {
	return []parserMethod{
		p.parseVerbPhrase,
		p.parseSuruVerb,
		p.parseNaAdjective,
		p.parseIAdjective,
		p.parseNounPhrase,
		p.parseParticlePhrase,
		p.parseCompoundNumber,
	}
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
	if k, ok := posMap[kind]; ok {
		t.POSMajor = k
		return nil
	}
	return fmt.Errorf("unknown POS1 %q", kind)
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

	lookup := start.BaseForm

	if checkIsGoDanVerbPotential(start) {
		lookup = processPotentialVerbBaseForm(start, p.lookupSet)
	}

	p.consumeSuffix(&surface)

	s := surface.String()

	if p.lookupSet[s] {
		return DisplayToken{
			Surface:  surface.String(),
			Lookup:   &s,
			POSMajor: tok.POSMajor,
		}, true
	}

	return DisplayToken{
		Surface:  surface.String(),
		Lookup:   &lookup,
		POSMajor: tok.POSMajor,
	}, true
}

func checkIsGoDanVerbPotential(rt RawToken) bool {
	runes := []rune(rt.Surface)
	if len(runes) < 2 {
		return false
	}
	if runes[len(runes)-1] != 'る' {
		return false
	}
	switch runes[len(runes)-2] {
	case 'せ', 'け', 'げ', 'べ', 'て', 'め', 'え', 'ね':
		return true
	default:
		return false
	}
}

// processPotentialVerbBaseForm checks if verb is in potential form and
// if so, returns infinitive of a verb as a baseForm
//
// Use cases:
//
// 1. rt is in potential form:
//
//	話せる -> 話す
//
// 2. rt looks like it is in potential form, but it isn't:
//
//	見せる -> 見せる
//
// 3. rt is a stem of potential form:
//
//	話せ -> 話す
func processPotentialVerbBaseForm(rt RawToken, lookupSet map[string]bool) string {
	initialBaseForm := rt.BaseForm
	if lookupSet[initialBaseForm] {
		return initialBaseForm
	}

	inf := getPotentialGoDanInfinitive(initialBaseForm)
	if lookupSet[inf] {
		return inf
	}

	return initialBaseForm
}

func getPotentialGoDanInfinitive(verb string) string {
	runes := []rune(verb)
	lv := len(runes)
	if lv < 2 {
		return verb
	}

	starting := string(runes[:lv-2])
	ending := string(runes[lv-2:])

	switch ending {
	case "せる":
		return starting + "す"
	case "ける":
		return starting + "く"
	case "げる":
		return starting + "ぐ"
	case "べる":
		return starting + "ぶ"
	case "てる":
		return starting + "つ"
	case "める":
		return starting + "む"
	case "える":
		return starting + "う"
	case "ねる":
		return starting + "ぬ"
	default:
		return verb
	}
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

	p.consumeSuffix(&surface)

	lookup := start.Surface
	return DisplayToken{
		Surface:  surface.String(),
		Lookup:   &lookup,
		POSMajor: tok.POSMajor,
	}, true
}

func (p *parser) consumeSuffix(surface *strings.Builder) {
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
		Surface:  surface.String(),
		Lookup:   &lookup,
		POSMajor: tok.POSMajor,
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
		Surface:  surface.String(),
		Lookup:   &lookup,
		POSMajor: tok.POSMajor,
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
			Surface:  s,
			Lookup:   &s,
			POSMajor: tok.POSMajor,
		}, true
	}

	if prefix, n := p.findLongestPrefix(nouns); n > 0 {
		p.pos = startPos + n
		return DisplayToken{
			Surface:  prefix,
			Lookup:   &prefix,
			POSMajor: tok.POSMajor,
		}, true
	}

	p.pos = startPos + 1
	s := tok.Surface
	return DisplayToken{
		Surface:  s,
		Lookup:   &s,
		POSMajor: tok.POSMajor,
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
	var b strings.Builder
	lengths := make([]int, 0, len(nouns))
	for _, n := range nouns {
		lengths = append(lengths, len(n.Surface))
		b.WriteString(n.Surface)
	}
	full := b.String()

	offset := len(full)
	for i := len(nouns); i >= 1; i-- {
		candidate := full[:offset]
		if p.lookupSet[candidate] {
			return candidate, i
		}
		offset -= lengths[i-1]
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
		Surface:  result,
		Lookup:   &result,
		POSMajor: tok.POSMajor,
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
		Surface:  result,
		Lookup:   &result,
		POSMajor: tok.POSMajor,
	}, true
}

// parseUnknown consumes a single token that was not recognized by any
// specialized parser.
func (p *parser) parseUnknown() DisplayToken {
	if p.eof() {
		return DisplayToken{}
	}

	tok := p.next()

	if tok.POSMajor == KindSymbol {
		return DisplayToken{}
	}

	s := tok.Surface

	lookup := &s
	if _, ok := p.lookupSet[s]; !ok {
		if _, ok := p.lookupSet[tok.BaseForm]; ok {
			lookup = &tok.BaseForm
		} else {
			lookup = nil
		}
	}

	return DisplayToken{
		Surface:  s,
		Lookup:   lookup,
		POSMajor: tok.POSMajor,
	}
}
