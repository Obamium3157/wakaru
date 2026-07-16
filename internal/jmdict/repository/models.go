package repository

type Entry struct {
	ID           string        `json:"id"`
	Kanji        []string      `json:"kanji"`
	Kana         []string      `json:"kana"`
	Translations []Translation `json:"translations"`
}

type Translation struct {
	SenseID int64   `json:"senseId"`
	Glosses []Gloss `json:"glosses"`
}

type Gloss struct {
	Lang string `json:"lang"`
	Text string `json:"text"`
}
