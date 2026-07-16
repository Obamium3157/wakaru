package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"wakaru/internal/examples"
	"wakaru/internal/jmdict/repository"
	"wakaru/internal/tokenize"

	"golang.org/x/sync/errgroup"
)

const maxAmountOfChannels = 5

type Result struct {
	Entries  []repository.Entry
	Examples []examples.Example
}

type Wakaru struct {
	db             *sql.DB
	repo           repository.Repository
	examplesClient *examples.Client

	lookupSet    map[string]bool
	httpSem      chan struct{}
	exampleCache sync.Map
}

func NewWakaru(ctx context.Context, sqlDriverName string, dbPath string) (*Wakaru, error) {
	db, err := openDB(sqlDriverName, dbPath)
	if err != nil {
		return nil, err
	}

	repo, err := createRepo(db)
	if err != nil {
		return nil, err
	}

	client := examples.NewClient()

	forms, err := repo.FindAllForms(ctx)
	if err != nil {
		return nil, err
	}

	lookupSet := makeLookupSet(forms)

	return &Wakaru{
		db:             db,
		repo:           repo,
		examplesClient: client,
		lookupSet:      lookupSet,
		httpSem:        make(chan struct{}, maxAmountOfChannels),
	}, nil
}

func (w *Wakaru) Run(ctx context.Context, input string) (string, []Result, error) {
	tokens, err := w.getDisplayTokens(input)
	if err != nil {
		return "", nil, err
	}

	httpTokens := 0
	for _, t := range tokens {
		if t.Lookup != nil {
			httpTokens++
		}
	}
	log.Printf("tokens: %d total, %d with lookup", len(tokens), httpTokens)

	results := make([]Result, len(tokens))
	g, ctx := errgroup.WithContext(ctx)

	for i, t := range tokens {
		g.Go(func() error {
			entries, err := w.FindEntries(ctx, t)
			if err != nil {
				return err
			}
			examples := w.FindExamples(ctx, t)
			results[i] = Result{entries, examples}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return "", nil, err
	}
	return formDisplaySearchString(tokens), results, nil
}

func (w *Wakaru) Close() error {
	return w.db.Close()
}

func (w *Wakaru) getDisplayTokens(input string) ([]tokenize.DisplayToken, error) {
	kagomeTokens, err := tokenize.Tokenize(input)
	if err != nil {
		return nil, err
	}

	displayTokens, err := tokenize.Parse(kagomeTokens, w.lookupSet)
	if err != nil {
		return nil, err
	}

	return displayTokens, nil
}

func (w *Wakaru) FindEntries(ctx context.Context, t tokenize.DisplayToken) ([]repository.Entry, error) {
	if t.Lookup == nil {
		return nil, nil
	}

	entries, err := w.repo.Find(ctx, *t.Lookup)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (w *Wakaru) FindExamples(ctx context.Context, t tokenize.DisplayToken) []examples.Example {
	if t.Lookup == nil {
		return nil
	}

	if cached, ok := w.exampleCache.Load(*t.Lookup); ok {
		return cached.([]examples.Example)
	}

	select {
	case w.httpSem <- struct{}{}:
		defer func() { <-w.httpSem }()
	case <-ctx.Done():
		return nil
	}

	if cached, ok := w.exampleCache.Load(*t.Lookup); ok {
		return cached.([]examples.Example)
	}

	examples, err := w.examplesClient.Search(ctx, examples.SearchParameters{
		Word:         *t.Lookup,
		MinWordCount: new(8),
		MaxWordCount: nil,
		Sort:         "relevance",
		Limit:        new(5),
	})
	if err != nil {
		log.Printf("tatoeba lookup failed for %q: %v", *t.Lookup, err)
		return nil
	}

	w.exampleCache.Store(*t.Lookup, examples)

	return examples
}

func openDB(driverName string, dbPath string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func createRepo(db *sql.DB) (repository.Repository, error) {
	repo, err := repository.NewSQLiteRepo(db)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func makeLookupSet(forms []string) map[string]bool {
	lookupSet := make(map[string]bool, len(forms))
	for _, f := range forms {
		lookupSet[f] = true
	}

	return lookupSet
}

func formDisplaySearchString(dt []tokenize.DisplayToken) string {
	var b strings.Builder
	for i, t := range dt {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(t.Surface)
	}
	return b.String()
}
