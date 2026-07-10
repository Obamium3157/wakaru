// Package jmdict converts JMDict from JSON file into database
// and provides API to access it
package jmdict

import (
	"database/sql"
	"database/sql/driver"
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
	Version       string         `json:"version"`
	Languages     []langCode     `json:"languages"`
	CommonOnly    bool           `json:"commonOnly"`
	DictDate      string         `json:"dictDate"`
	DictRevisions []string       `json:"dictRevisions"`
	Tags          map[tag]string `json:"tags"`
	Words         []Word         `json:"words"`
}

type Word struct {
	ID     string  `json:"id"`
	Kana   []kana  `json:"kana"`
	Kanji  []kanji `json:"kanji"`
	Senses []Sense `json:"sense"`
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
	Headword   string
	Reading    sql.NullString
	SenseIndex sql.NullInt64
}

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
		return x.initWord(raw)
	case 2:
		return x.initWordIndexOrReading(raw)
	case 3:
		return x.initWordReadingIndex(raw)
	default:
		return fmt.Errorf("invalid xref length: %d", len(raw))
	}
}

func (x *xref) initWord(raw []json.RawMessage) error {
	headword, err := GetStr(raw, 0)
	if err != nil {
		return err
	}

	x.Headword = headword

	return nil
}

func (x *xref) initWordIndexOrReading(raw []json.RawMessage) error {
	headword, err := GetStr(raw, 0)
	if err != nil {
		return err
	}

	x.Headword = headword

	if idx, isInt, err := GetInt(raw, 1); err != nil {
		return err
	} else if isInt {
		x.SenseIndex = sql.NullInt64{
			Int64: int64(idx),
			Valid: true,
		}
		return nil
	}

	reading, err := GetStr(raw, 1)
	if err != nil {
		return err
	}

	x.Reading = sql.NullString{
		String: reading,
		Valid:  true,
	}

	return nil
}

func (x *xref) initWordReadingIndex(raw []json.RawMessage) error {
	headword, err := GetStr(raw, 0)
	if err != nil {
		return err
	}

	reading, err := GetStr(raw, 1)
	if err != nil {
		return err
	}

	index, _, err := GetInt(raw, 2)
	if err != nil {
		return err
	}

	x.Headword = headword
	x.Reading = sql.NullString{
		String: reading,
		Valid:  true,
	}
	x.SenseIndex = sql.NullInt64{
		Int64: int64(index),
		Valid: true,
	}

	return nil
}

func (x xref) HasReading() bool {
	return x.Reading.Valid
}

func (x xref) HasSenseIndex() bool {
	return x.SenseIndex.Valid
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

func (g gender) Value() (driver.Value, error) {
	if !g.Valid {
		return nil, nil
	}

	return g.String, nil
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

func (g glossType) Value() (driver.Value, error) {
	if !g.Valid {
		return nil, nil
	}

	return g.String, nil
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

func (t text) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}

	return t.String, nil
}

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
