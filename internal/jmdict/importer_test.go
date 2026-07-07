package jmdict

import (
	"testing"
)

func tryConstructXref[T xrefPayload](val T, t *testing.T) {
	xref := newXref(val)
	if xref == nil {
		t.Error("could not create xref")
	}
}

func TestConstructXref(t *testing.T) {
	tryConstructXref(xrefWordReadingIndex{Kanji: "豚", Kana: "ぶた", SenseIndex: 1}, t)
	tryConstructXref(xrefWordReading{Kanji: "豚", Kana: "ぶた"}, t)
	tryConstructXref(xrefWordIndex{KanjiOrKana: "豚", SenseIndex: 1}, t)
	tryConstructXref(xrefWord{KanjiOrKana: "豚"}, t)
}

func tryMakeGender(s string, t *testing.T) {
	gender, err := makeGender(s)
	if err != nil || gender == "" {
		t.Errorf("could not make gender: %v", err)
	}
}

func TestMakeGender(t *testing.T) {
	tryMakeGender("masculine", t)
	tryMakeGender("feminine", t)
	tryMakeGender("neuter", t)

	gender, err := makeGender("alien")
	if err == nil && gender == "" {
		t.Error("created wrong gender (alien)")
	}
}

func tryMakeGlossType(s string, t *testing.T) {
	gloss, err := makeGlossType(s)
	if err != nil || gloss == "" {
		t.Errorf("could not make gloss type: %v", err)
	}
}

func TestMakeGlossType(t *testing.T) {
	tryMakeGlossType("literal", t)
	tryMakeGlossType("figurative", t)
	tryMakeGlossType("explanation", t)
	tryMakeGlossType("trademark", t)

	gloss, err := makeGlossType("nonsense")
	if err == nil && gloss == "" {
		t.Error("created wrong gloss type (nonsense)")
	}
}
