package repository

type RubySegment struct {
	Text    string  `json:"text"`
	Reading *string `json:"reading,omitempty"`
}

type Entry struct {
	ID           string        `json:"id"`
	Kanji        []string      `json:"kanji"`
	Kana         []string      `json:"kana"`
	Ruby         []RubySegment `json:"ruby,omitempty"`
	Translations []Translation `json:"translations"`
}

type Translation struct {
	SenseID int64   `json:"senseId"`
	Pos     *string `json:"pos,omitempty"`
	Glosses []Gloss `json:"glosses"`
}

type Gloss struct {
	Lang string `json:"lang"`
	Text string `json:"text"`
}
