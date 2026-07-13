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

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <word>", os.Args[0])
	}
	word := os.Args[1]

	db := mustOpenDB()
	defer func() {
		_ = db.Close()
	}()

	repo := createRepo(db)

	entries, err := repo.Find(word)
	if err != nil {
		log.Fatal(err)
	}

	printEntries(entries)

	examples, err := examples.Search(word)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Examples: ")
	printExamples(examples)
}

func mustOpenDB() *sql.DB {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		log.Fatal(err)
	}

	return db
}

func createRepo(db *sql.DB) repository.Repository {
	repo, err := repository.NewSQLiteRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	return repo
}

func printEntries(entries []repository.Entry) {
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

func printExamples(examples []examples.Example) {
	for idx, example := range examples {
		fmt.Printf("  Example #%d: %s\n", idx, example.Text)
	}
}
