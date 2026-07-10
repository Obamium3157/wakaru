package repository

type Entry struct {
	ID           string
	Kanji        []string
	Kana         []string
	Translations []Translation
}

type Translation struct {
	SenseID int64
	Glosses []Gloss
}

type Gloss struct {
	Lang string
	Text string
}
