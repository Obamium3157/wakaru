package main

import (
	"fmt"
	"log"
	"os"

	"wakaru/internal/examples"
	"wakaru/internal/jmdict/repository"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	mustLoadEnv()

	input := mustGetInputFromArgs()

	w, err := NewWakaru("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	results, err := w.Run(input)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range results {
		printEntries(r.Entries)
		printExamples(r.Examples)
	}
}

func mustLoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}

func mustGetInputFromArgs() string {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <word>", os.Args[0])
	}
	return os.Args[1]
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
