package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"wakaru/internal/jmdict"
	"wakaru/internal/jmdict/repository"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("got an error trying to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("got an error trying to close database: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}
	fmt.Println("Successfully connected to database")

	repo, err := repository.NewSQLiteRepo(db)
	if err != nil {
		log.Fatalf("error creating jmdict repo: %v", err)
	}

	entries, err := repo.Find("猫")
	if err != nil {
		log.Fatal(err)
	}

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

func LoadDictFromJSON(jsonPath string) (*jmdict.Dictionary, error) {
	dict, err := jmdict.InitDictionary(jsonPath)
	if err != nil {
		return nil, err
	}

	return dict, nil
}
