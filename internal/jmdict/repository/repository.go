package repository

import "context"

type Repository interface {
	Find(ctx context.Context, text string) ([]Entry, error)

	FindFiltered(ctx context.Context, text string, posTags []string) ([]Entry, error)

	FindByKanji(ctx context.Context, text string) ([]Entry, error)

	FindByKana(ctx context.Context, text string) ([]Entry, error)

	FindAllForms(ctx context.Context) ([]string, error)
}
