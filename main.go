package main

import (
	"log"
	"os"

	"wakaru/internal/tokenize"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <word>", os.Args[0])
	}
	input := os.Args[1]

	w, err := NewWakaru()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = w.Close()
	}()

	displayTokens, err := w.GetDisplayTokens(input)
	if err != nil {
		log.Fatal(err)
	}

	if err := run(w, displayTokens); err != nil {
		log.Fatal(err)
	}
}

func run(w *Wakaru, ts []tokenize.DisplayToken) error {
	for _, t := range ts {
		entries, err := w.FindEntries(t)
		if err != nil {
			return err
		}
		PrintEntries(entries)

		examples, err := w.FindExamples(t)
		if err != nil {
			return err
		}
		PrintExamples(examples)
	}

	return nil
}
