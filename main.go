package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"wakaru/internal/jmdict"

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

	/*
			* `id` INTEGER PRIMARY KEY,
		  `word_id` TEXT NOT NULL REFERENCES `word` (`id`) ON DELETE CASCADE,
		  `is_common` INTEGER NOT NULL DEFAULT 0 CHECK (`is_common` IN (0, 1)),
		  `text` TEXT NOT NULL,
		  `display_order` INTEGER NOT NULL
	*/

	// if err := jmdict.InitJMDictDB(db); err != nil {
	// 	log.Fatalf("failed to init jmdict database: %v", err)
	// }
	//
	// dict, err := LoadDictFromJSON("../jmdict/jmdict-eng-3.6.2.json")
	// if err != nil {
	// 	log.Fatalf("failed to lead dictionary: %v", err)
	// }
	//
	// if err := jmdict.FillDatabase(db, dict); err != nil {
	// 	log.Fatalf("failed to fill jmdict database: %v", err)
	// }
}

func LoadDictFromJSON(jsonPath string) (*jmdict.Dictionary, error) {
	dict, err := jmdict.InitDictionary(jsonPath)
	if err != nil {
		return nil, err
	}

	return dict, nil
}

func PrintSense(sense jmdict.Sense) {
	fmt.Println("\tAntonym: ", sense.Antonym)
	fmt.Println("\tApplies to kana: ", sense.AppliesToKana)
	fmt.Println("\tApplies to kanji: ", sense.AppliesToKanji)
	fmt.Println("\tDialect: ", sense.Dialect)
	fmt.Println("\tField: ", sense.Field)
	fmt.Println("\tGloss: ", sense.Gloss)
	fmt.Println("\tInfo: ", sense.Info)
	fmt.Println("\tLanguage source: ", sense.LanguageSource)
	fmt.Println("\tMisc: ", sense.Misc)
	fmt.Println("\tPart of speech: ", sense.PartOfSpeech)
	fmt.Println("\tRelated: ", sense.Related)
}

func PrintWord(word jmdict.Word) {
	fmt.Println("ID: ", word.ID)
	fmt.Println("Kana :", word.Kana)
	fmt.Println("Kanji: ", word.Kanji)
	fmt.Println("Senses: ")
	for _, sense := range word.Senses {
		PrintSense(sense)
	}
}
