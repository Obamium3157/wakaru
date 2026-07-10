package jmdict

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tryMakeGender(s string, t *testing.T) {
	gender, err := makeGender(s)
	if err != nil || !gender.IsValid() {
		t.Errorf("could not make gender: %v", err)
	}
}

func TestMakeGender(t *testing.T) {
	tryMakeGender("masculine", t)
	tryMakeGender("feminine", t)
	tryMakeGender("neuter", t)

	gender, err := makeGender("alien")
	if err == nil && !gender.IsValid() {
		t.Error("created wrong gender (alien)")
	}
}

func tryMakeGlossType(s string, t *testing.T) {
	gloss, err := makeGlossType(s)
	if err != nil || gloss.String == "" {
		t.Errorf("could not make gloss type: %v", err)
	}
}

func TestMakeGlossType(t *testing.T) {
	tryMakeGlossType("literal", t)
	tryMakeGlossType("figurative", t)
	tryMakeGlossType("explanation", t)
	tryMakeGlossType("trademark", t)

	gloss, err := makeGlossType("nonsense")
	if err == nil && gloss.String == "" {
		t.Error("created wrong gloss type (nonsense)")
	}
}

func TestUnmarshalXrefWord(t *testing.T) {
	data := []byte(`["一の字点"]`)
	var xref xref

	if err := xref.UnmarshalJSON(data); err != nil {
		t.Errorf("got error trying to parse JSON: %v", err)
	}

	assert.Equal(t, xref.Headword, "一の字点")
}

func TestUnmarshalXrefWordIndex(t *testing.T) {
	data := []byte(`["〇〇", 1]`)
	var xref xref

	if err := xref.UnmarshalJSON(data); err != nil {
		t.Errorf("got error trying to parse JSON: %v", err)
	}

	assert.Equal(t, xref.Headword, "〇〇")
	assert.Equal(t, xref.SenseIndex.Int64, int64(1))
}

func TestUnmarshalXrefWordReading(t *testing.T) {
	data := []byte(`["丸", "まる"]`)
	var xref xref

	if err := xref.UnmarshalJSON(data); err != nil {
		t.Errorf("got error trying to parse JSON: %v", err)
	}

	assert.Equal(t, xref.Headword, "丸")
	assert.Equal(t, xref.Reading.String, "まる")
}

func TestUnmarshalXrefWordReadingIndex(t *testing.T) {
	data := []byte(`["丸", "まる・1", 1]`)
	var xref xref

	if err := xref.UnmarshalJSON(data); err != nil {
		t.Errorf("got error trying to parse JSON: %v", err)
	}

	assert.Equal(t, xref.Headword, "丸")
	assert.Equal(t, xref.Reading, "まる・1")
	assert.Equal(t, xref.SenseIndex, 1)
}

func testGender(genderStr string, t *testing.T) {
	jsonStr := fmt.Sprintf(`"%s"`, genderStr)
	data := []byte(jsonStr)

	var g gender
	if err := json.Unmarshal(data, &g); err != nil {
		t.Errorf("got error trying to parse JSON for %q: %v", genderStr, err)
		return
	}
	if g.String != genderStr || !g.Valid {
		t.Errorf("gender should be %v; got (%v %v)", genderStr, g.String, g.Valid)
	}
}

func TestUnmarshalGender(t *testing.T) {
	data := []byte(`null`)
	var g gender
	if err := json.Unmarshal(data, &g); err != nil {
		t.Errorf("got error trying to parse JSON: %v", err)
	}
	if g.String != "" || g.Valid {
		t.Errorf("g.String should be \"\" and g.Valid should be false (got (%v, %v))", g.String, g.Valid)
	}

	testGender("masculine", t)
	testGender("feminine", t)
	testGender("neuter", t)
}

func testGlossType(gtStr string, t *testing.T) {
	jsonStr := fmt.Sprintf(`"%s"`, gtStr)
	data := []byte(jsonStr)

	var g glossType
	if err := json.Unmarshal(data, &g); err != nil {
		t.Errorf("got error trying to parse JSON for %q: %v", gtStr, err)
		return
	}

	if g.String != gtStr || !g.Valid {
		t.Errorf("gloss type should be %v; got (%v %v)", gtStr, g.String, g.Valid)
	}
}

func TestUnmarshalGlossType(t *testing.T) {
	data := []byte(`null`)
	var g glossType
	if err := json.Unmarshal(data, &g); err != nil {
		t.Errorf("got error trying to parse JSON: %v", err)
	}
	if g.String != "" || g.Valid {
		t.Errorf("g.String should be \"\" and g.Valid should be false (got (%v, %v))", g.String, g.Valid)
	}

	testGlossType("literal", t)
	testGlossType("figurative", t)
	testGlossType("explanation", t)
	testGlossType("trademark", t)
}
