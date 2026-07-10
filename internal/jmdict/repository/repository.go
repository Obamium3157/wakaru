package repository

type Repository interface {
	Find(text string) ([]Entry, error)

	FindByKanji(text string) ([]Entry, error)

	FindByKana(text string) ([]Entry, error)
}
