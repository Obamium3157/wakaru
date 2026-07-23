// Package wakaru provides Japanese text analysis by combining morphological
// tokenization using Kagome, JMDict dictionary lookups, and example sentence
// retrieval from the Tatoeba API.
package wakaru

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"wakaru/internal/anki"
	"wakaru/internal/examples"
	"wakaru/internal/jmdict/repository"
	"wakaru/internal/tokenize"

	"golang.org/x/sync/errgroup"
)

const maxAmountOfChannels = 5

var kindToPOSTags = map[tokenize.Kind][]string{
	tokenize.KindNoun: {
		"n", "n-t", "n-adv", "n-pref", "n-suf", "n-pr", "pn",
	},
	tokenize.KindVerb: {
		"v1", "v1-s", "v5aru", "v5b", "v5g", "v5k", "v5k-s",
		"v5m", "v5n", "v5r", "v5r-i", "v5s", "v5t", "v5u",
		"v5u-s", "v5uru", "v2a-s", "v2b-k", "v2b-s", "v2d-k",
		"v2d-s", "v2g-k", "v2g-s", "v2h-k", "v2h-s", "v2k-k",
		"v2k-s", "v2m-k", "v2m-s", "v2n-s", "v2r-k", "v2r-s",
		"v2s-s", "v2t-k", "v2t-s", "v2w-s", "v2y-k", "v2y-s",
		"v2z-s", "v4b", "v4g", "v4h", "v4k", "v4m", "v4n",
		"v4r", "v4s", "v4t", "vk", "vn", "vr", "vs", "vs-c",
		"vs-i", "vs-s", "v-unspec", "vz", "vi", "vt",
	},
	tokenize.KindAuxVerb:  {"aux-v", "aux"},
	tokenize.KindParticle: {"prt"},
	tokenize.KindPrefix:   {"pref", "n-pref"},
	tokenize.KindAdjective: {
		"adj-i", "adj-ix", "adj-na", "adj-no", "adj-pn", "adj-f",
		"adj-t", "adj-kari", "adj-ku", "adj-shiku", "adj-nari",
		"aux-adj",
	},
	tokenize.KindConjunction:  {"conj"},
	tokenize.KindInterjection: {"int"},
	tokenize.KindAdverb:       {"adv", "adv-to", "n-adv"},
	tokenize.KindPrenominal:   {"adj-pn", "adj-f"},
	tokenize.KindOther:        {"unc", "oth"},
}

type Result struct {
	Entries  []repository.Entry `json:"entries"`
	Examples []examples.Example `json:"examples"`
	PosMajor string             `json:"posMajor"`
}

type Wakaru struct {
	db             *sql.DB
	repo           repository.Repository
	examplesClient *examples.Client
	ankiClient     *anki.Client

	lookupSet    map[string]bool
	httpSem      chan struct{}
	exampleCache sync.Map
}

func NewWakaru(ctx context.Context, sqlDriverName string, dbPath string, ankiPort int) (*Wakaru, error) {
	db, err := openDB(sqlDriverName, dbPath)
	if err != nil {
		return nil, err
	}

	repo, err := createRepo(db)
	if err != nil {
		return nil, err
	}

	examplesClient := examples.NewClient()

	ankiClient, err := anki.NewClient(ankiPort)
	if err != nil {
		return nil, err
	}

	forms, err := repo.FindAllForms(ctx)
	if err != nil {
		return nil, err
	}

	lookupSet := makeLookupSet(forms)

	return &Wakaru{
		db:             db,
		repo:           repo,
		examplesClient: examplesClient,
		ankiClient:     ankiClient,
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
			results[i] = Result{
				Entries:  entries,
				Examples: examples,
				PosMajor: t.POSMajor.String(),
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return "", nil, err
	}
	return formDisplaySearchString(tokens), results, nil
}

func (w *Wakaru) FindWord(ctx context.Context, text string, posMajor string) ([]repository.Entry, error) {
	if posMajor != "" {
		if kind, ok := tokenize.ParseKind(posMajor); ok {
			if tags, ok := kindToPOSTags[kind]; ok {
				entries, err := w.repo.FindFiltered(ctx, text, tags)
				if err != nil {
					return nil, err
				}
				if len(entries) > 0 {
					return entries, nil
				}
			}
		}
	}
	return w.repo.Find(ctx, text)
}

func (w *Wakaru) AddBasicNote(ctx context.Context, deckName, front, back string, tags []string) error {
	if err := w.ankiClient.AddBasicNote(ctx, anki.AddBasicNoteRequest{
		DeckName: deckName,
		Front:    front,
		Back:     back,
		Tags:     tags,
	}); err != nil {
		return err
	}

	return nil
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

	if tags, ok := kindToPOSTags[t.POSMajor]; ok {
		entries, err := w.repo.FindFiltered(ctx, *t.Lookup, tags)
		if err != nil {
			return nil, err
		}
		if len(entries) > 0 {
			return entries, nil
		}
	}

	return w.repo.Find(ctx, *t.Lookup)
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
		Limit:        new(10),
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
