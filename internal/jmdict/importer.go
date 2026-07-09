// Package jmdict converts JMDict from JSON file into database
// and provides API to access it
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
	Words         []Word            `json:"words"`
}

type Word struct {
	ID    string  `json:"id"`
	Kana  []kana  `json:"kana"`
	Kanji []kanji `json:"kanji"`
	Sense []Sense `json:"sense"`
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

type Sense struct {
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

// TODO: написать негативные тесты для UnmarshalJSON
func (x *xref) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch len(raw) {
	case 1:
		if err := initXrefWord(x, raw); err != nil {
			return err
		}
	case 2:
		if err := initXrefWOrdIndexOrWordReading(x, raw); err != nil {
			return err
		}
	case 3:
		if err := initXrefWordReadingIndex(x, raw); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid xref length: %d", len(raw))
	}

	return nil
}

func initXrefWord(xref *xref, raw []json.RawMessage) error {
	s, err := GetStr(raw, 0)
	if err != nil {
		return err
	}
	xref.Value = xrefWord{KanjiOrKana: s}

	return nil
}

func initXrefWOrdIndexOrWordReading(xref *xref, raw []json.RawMessage) error {
	s0, err := GetStr(raw, 0)
	if err != nil {
		return err
	}
	idx, isInt, err := GetInt(raw, 1)
	if err != nil {
		return err
	}

	if isInt {
		xref.Value = xrefWordIndex{KanjiOrKana: s0, SenseIndex: idx}
	} else {
		s1, err := GetStr(raw, 1)
		if err != nil {
			return err
		}
		xref.Value = xrefWordReading{Kanji: s0, Kana: s1}
	}

	return nil
}

func initXrefWordReadingIndex(xref *xref, raw []json.RawMessage) error {
	s0, err := GetStr(raw, 0)
	if err != nil {
		return err
	}
	s1, err := GetStr(raw, 1)
	if err != nil {
		return err
	}
	idx, _, err := GetInt(raw, 2)
	if err != nil {
		return err
	}
	xref.Value = xrefWordReadingIndex{Kanji: s0, Kana: s1, SenseIndex: idx}

	return nil
}

type gloss struct {
	Gender *gender    `json:"gender"`
	Lang   langCode   `json:"lang"`
	Text   string     `json:"text"`
	Type   *glossType `json:"type"`
}

type gender sql.NullString

func (g gender) IsValid() bool {
	switch g.String {
	case GenderMasculine, GenderFeminine, GenderNeuter:
		return true
	}

	return false
}

func makeGender(s string) (*gender, error) {
	g := gender{String: s, Valid: true}
	if !g.IsValid() {
		return nil, fmt.Errorf("invalid gender value: %v", s)
	}

	return &g, nil
}

func (g *gender) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		g.String = ""
		g.Valid = false
		return nil
	}
	val, err := makeGender(s)
	if err != nil {
		g.String = ""
		g.Valid = false
		return err
	}
	*g = *val
	return nil
}

type langCode string

type glossType sql.NullString

func (g glossType) IsValid() bool {
	switch g.String {
	case GlossLiteral, GlossFigurative, GlossExplanation, GlossTrademart:
		return true
	}

	return false
}

func makeGlossType(s string) (*glossType, error) {
	g := glossType{String: s, Valid: true}
	if !g.IsValid() {
		return nil, fmt.Errorf("invalid glossType value: %v", s)
	}

	return &g, nil
}

func (g *glossType) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		g.String = ""
		g.Valid = false

		return err
	}

	val, err := makeGlossType(s)
	if err != nil {
		g.String = ""
		g.Valid = false

		return err
	}

	*g = *val
	return nil
}

type languageSource struct {
	Full  bool     `json:"full"`
	Lang  langCode `json:"lang"`
	Text  text     `json:"text"`
	Wasei bool     `json:"wasei"`
}

type text sql.NullString

func (t *text) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		t.String = ""
		t.Valid = false
		return err
	}

	t.String = s
	t.Valid = true

	return nil
}

func InitDictionary(jsonFilename string) (*Dictionary, error) {
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
