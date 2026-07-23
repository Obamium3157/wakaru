package anki

type AddBasicNoteRequest struct {
	DeckName string
	Front    string
	Back     string
	Tags     []string
}
