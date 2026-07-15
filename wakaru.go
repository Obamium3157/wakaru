package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"wakaru/internal/examples"
	"wakaru/internal/jmdict/repository"
	"wakaru/internal/tokenize"

	"github.com/joho/godotenv"
)

type Wakaru struct {
	db             *sql.DB
	repo           repository.Repository
	examplesClient *examples.Client

	lookupSet map[string]bool
}

func NewWakaru() (*Wakaru, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}

	repo := createRepo(db)

	client := examples.NewClient()

	forms, err := repo.FindAllForms()
	if err != nil {
		return nil, err
	}

	lookupSet := makeLookupSet(forms)

	return &Wakaru{
		db:             db,
		repo:           repo,
		examplesClient: client,
		lookupSet:      lookupSet,
	}, nil
}

func (w *Wakaru) Close() error {
	return w.db.Close()
}

func PrintEntries(entries []repository.Entry) {
	for _, entry := range entries {
		fmt.Printf("ID: %s\n", entry.ID)

		fmt.Println("Kanji:")
		for _, k := range entry.Kanji {
			fmt.Printf("  %s\n", k)
		}

		fmt.Println("Kana:")
		for _, k := range entry.Kana {
			fmt.Printf("  %s\n", k)
		}

		fmt.Println("Translations:")
		for _, t := range entry.Translations {
			fmt.Printf("  Sense %d\n", t.SenseID)

			for _, g := range t.Glosses {
				fmt.Printf("    [%s] %s\n", g.Lang, g.Text)
			}
		}

		fmt.Println()
	}
}

func PrintExamples(examples []examples.Example) {
	for idx, example := range examples {
		fmt.Printf("  Example #%d: %s\n", idx, example.Text)
	}
}

func (w *Wakaru) GetDisplayTokens(input string) ([]tokenize.DisplayToken, error) {
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

func (w *Wakaru) FindEntries(t tokenize.DisplayToken) ([]repository.Entry, error) {
	if t.Lookup == nil {
		return nil, nil
	}

	entries, err := w.repo.Find(*t.Lookup)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (w *Wakaru) FindExamples(t tokenize.DisplayToken) ([]examples.Example, error) {
	if t.Lookup == nil {
		return nil, nil
	}

	examples, err := w.examplesClient.Search(*t.Lookup)
	if err != nil {
		return nil, err
	}

	return examples, nil
}

func openDB() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
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

func createRepo(db *sql.DB) *repository.SQLiteRepo {
	repo, err := repository.NewSQLiteRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	return repo
}

func makeLookupSet(forms []string) map[string]bool {
	lookupSet := make(map[string]bool, len(forms))
	for _, f := range forms {
		lookupSet[f] = true
	}

	return lookupSet
}
