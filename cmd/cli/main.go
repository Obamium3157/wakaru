package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"wakaru/internal/examples"
	"wakaru/internal/jmdict/repository"
	"wakaru/internal/wakaru"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	mustLoadEnv()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	input := mustGetInputFromArgs()

	ankiPort, err := strconv.Atoi(os.Getenv("ANKI_PORT"))
	if err != nil {
		log.Fatalf("failed parsing anki port: %v", err)
	}

	w, err := wakaru.NewWakaru(
		ctx,
		"sqlite3",
		os.Getenv("DB_PATH"),
		os.Getenv("JMDICT_PATH"),
		ankiPort,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	start := time.Now()
	dispStr, results, err := w.RunSynchronous(ctx, input)
	elapsed := time.Since(start)
	if err != nil {
		if ctx.Err() != nil {
			os.Exit(0)
		}
		log.Fatal(err)
	}
	log.Printf("done in %v", elapsed)

	fmt.Println(dispStr, ": ")
	for _, r := range results {
		printEntries(r.Entries)
		printExamples(r.Examples)
		fmt.Println("------------------------------")
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
	if len(entries) == 0 {
		log.Println("empty entries array")
		return
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

func printExamples(examples []examples.Example) {
	if len(examples) == 0 {
		log.Println("empty examples array")
		return
	}
	for idx, example := range examples {
		fmt.Printf("  Example #%d: %s\n", idx+1, example.Text)
	}
}
