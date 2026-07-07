package jmdict

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
)

const (
	GenderMasculine = "masculine"
	GenderFeminine  = "feminine"
	GenderNeuter    = "neuter"

	GlossLiteral     = "literal"
	GlossFigurative  = "figurative"
	GlossExplanation = "explanation"
	GlossTrademart   = "trademark"
)

type Dictionary struct {
	Version       string            `json:"version"`
	Languages     []langCode        `json:"languages"`
	CommonOnly    bool              `json:"commonOnly"`
	DictDate      string            `json:"dictDate"`
	DictRevisions []string          `json:"dictRevisions"`
	Tags          map[string]string `json:"tags"`
	Words         []word            `json:"words"`
}

type word struct {
	ID    string  `json:"id"`
	Kana  []kana  `json:"kana"`
	Kanji []kanji `json:"kanji"`
	Sense []sense `json:"sense"`
}

type kana struct {
	AppliesToKanji []string `json:"appliesToKanji"`
	Common         bool     `json:"common"`
	Tags           []tag    `json:"tags"`
	Text           string   `json:"text"`
}

type kanji struct {
	Common bool   `json:"common"`
	Tags   []tag  `json:"tags"`
	Text   string `json:"text"`
}

type sense struct {
	Antonym        []xref           `json:"antonym"`
	AppliesToKana  []string         `json:"appliesToKana"`
	AppliesToKanji []string         `json:"appliesToKanji"`
	Dialect        []tag            `json:"dialect"`
	Field          []tag            `json:"field"`
	Gloss          []gloss          `json:"gloss"`
	Info           []string         `json:"info"`
	LanguageSource []languageSource `json:"languageSource"`
	Misc           []tag            `json:"misc"`
	PartOfSpeech   []tag            `json:"partOfSpeech"`
	Related        []xref           `json:"related"`
}

type tag string

type xref struct {
	Value any
}

type xrefPayload interface {
	xrefWordReadingIndex | xrefWordReading | xrefWordIndex | xrefWord
}

type xrefWordReadingIndex struct {
	Kanji      string
	Kana       string
	SenseIndex int
}

type xrefWordReading struct {
	Kanji string
	Kana  string
}

type xrefWordIndex struct {
	KanjiOrKana string
	SenseIndex  int
}

type xrefWord struct {
	KanjiOrKana string
}

func newXref[T xrefPayload](val T) *xref {
	return &xref{
		Value: val,
	}
}

type gloss struct {
	Gender *gender    `json:"gender"`
	Lang   langCode   `json:"lang"`
	Text   string     `json:"text"`
	Type   *glossType `json:"type"`
}

type gender string

func (g gender) IsValid() bool {
	switch g {
	case GenderMasculine, GenderFeminine, GenderNeuter:
		return true
	}

	return false
}

func makeGender(s string) (gender, error) {
	g := gender(s)
	if !g.IsValid() {
		return "", fmt.Errorf("invalid gender value: %v", s)
	}

	return g, nil
}

type langCode string

type glossType string

func (g glossType) IsValid() bool {
	switch g {
	case GlossLiteral, GlossFigurative, GlossExplanation, GlossTrademart:
		return true
	}

	return false
}

func makeGlossType(s string) (glossType, error) {
	g := glossType(s)
	if !g.IsValid() {
		return "", fmt.Errorf("invalid glossType value: %v", s)
	}

	return g, nil
}

type languageSource struct {
	Full  bool
	Lang  langCode
	text  sql.NullString
	wasei bool
}

func InitDB(jsonFilename string) (*Dictionary, error) {
	fileBytes, err := os.ReadFile(jsonFilename)
	if err != nil {
		return nil, err
	}

	var dict Dictionary
	err = json.Unmarshal(fileBytes, &dict)
	if err != nil {
		return nil, err
	}

	return &dict, nil
}
