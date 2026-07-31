// Package repository provides API for accessing JMDict Database
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"wakaru/internal/sqlutils"
)

type SQLiteRepo struct {
	db *sql.DB

	find *sql.Stmt

	findByKanji *sql.Stmt
	findByKana  *sql.Stmt

	findAllForms *sql.Stmt
}

func NewSQLiteRepo(db *sql.DB) (*SQLiteRepo, error) {
	repo := &SQLiteRepo{db: db}

	var err error

	if repo.find, err = db.Prepare(findQuery); err != nil {
		return nil, err
	}
	if repo.findByKanji, err = db.Prepare(findByKanjiQuery); err != nil {
		return nil, err
	}
	if repo.findByKana, err = db.Prepare(findByKanaQuery); err != nil {
		return nil, err
	}
	if repo.findAllForms, err = db.Prepare(findAllFormsQuery); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteRepo) FindByKanji(ctx context.Context, text string) ([]Entry, error) {
	rows, err := r.findByKanji.QueryContext(ctx, text)
	if err != nil {
		return nil, err
	}

	return r.loadEntries(ctx, rows)
}

func (r *SQLiteRepo) FindByKana(ctx context.Context, text string) ([]Entry, error) {
	rows, err := r.findByKana.QueryContext(ctx, text)
	if err != nil {
		return nil, err
	}

	return r.loadEntries(ctx, rows)
}

func (r *SQLiteRepo) Find(ctx context.Context, text string) ([]Entry, error) {
	rows, err := r.find.QueryContext(ctx, text, text)
	if err != nil {
		return nil, err
	}

	return r.loadEntries(ctx, rows)
}

func (r *SQLiteRepo) FindFiltered(ctx context.Context, text string, posTags []string) ([]Entry, error) {
	query := fmt.Sprintf(
		findFilteredQuery,
		strings.Join(
			sqlutils.GetPlaceholders(len(posTags)),
			",",
		),
	)

	args := make([]any, 0, 2+len(posTags))
	args = append(args, text, text)
	for _, tag := range posTags {
		args = append(args, tag)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return r.loadEntries(ctx, rows)
}

func (r *SQLiteRepo) FindAllForms(ctx context.Context) ([]string, error) {
	rows, err := r.findAllForms.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var forms []string

	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		forms = append(forms, text)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return forms, nil
}

type entryData struct {
	kanji         map[string][]string
	kana          map[string][]string
	translations  map[string][]Translation
	kanjiReadings map[string]map[string]string
}

func (r *SQLiteRepo) loadEntries(ctx context.Context, rows *sql.Rows) ([]Entry, error) {
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	wordIDs, err := scanWordIDs(rows)
	if err != nil {
		return nil, err
	}

	if len(wordIDs) == 0 {
		return nil, nil
	}

	data, err := r.loadEntryData(ctx, wordIDs)
	if err != nil {
		return nil, err
	}

	return buildEntries(wordIDs, data), nil
}

func scanWordIDs(rows *sql.Rows) ([]string, error) {
	var wordIDs []string

	for rows.Next() {
		var wordID string

		if err := rows.Scan(&wordID); err != nil {
			return nil, err
		}

		wordIDs = append(wordIDs, wordID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wordIDs, nil
}

func (r *SQLiteRepo) loadEntryData(ctx context.Context, wordIDs []string) (entryData, error) {
	kanji, err := r.loadKanjiBatch(ctx, wordIDs)
	if err != nil {
		return entryData{}, err
	}

	kana, err := r.loadKanaBatch(ctx, wordIDs)
	if err != nil {
		return entryData{}, err
	}

	translations, err := r.loadTranslationsBatch(ctx, wordIDs)
	if err != nil {
		return entryData{}, err
	}

	kanjiReadings, err := r.loadKanaReadingsForKanjiBatch(ctx, wordIDs)
	if err != nil {
		return entryData{}, err
	}

	return entryData{
		kanji:         kanji,
		kana:          kana,
		translations:  translations,
		kanjiReadings: kanjiReadings,
	}, nil
}

func buildEntries(wordIDs []string, data entryData) []Entry {
	entries := make([]Entry, 0, len(wordIDs))

	for _, wordID := range wordIDs {
		entryKanji := data.kanji[wordID]
		entryKana := data.kana[wordID]

		entries = append(entries, Entry{
			ID:           wordID,
			Kanji:        entryKanji,
			Kana:         entryKana,
			Ruby:         buildEntryRuby(entryKanji, entryKana, data.kanjiReadings[wordID]),
			Translations: data.translations[wordID],
		})
	}

	return entries
}

func (r *SQLiteRepo) loadKanjiBatch(ctx context.Context, wordIDs []string) (map[string][]string, error) {
	query := fmt.Sprintf(
		findKanjiBatchQuery,
		strings.Join(sqlutils.GetPlaceholders(len(wordIDs)), ","),
	)

	args := make([]any, len(wordIDs))
	for i, id := range wordIDs {
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	result := make(map[string][]string, len(wordIDs))

	for rows.Next() {
		var wordID, text string

		if err := rows.Scan(&wordID, &text); err != nil {
			return nil, err
		}

		result[wordID] = append(result[wordID], text)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *SQLiteRepo) loadKanaBatch(ctx context.Context, wordIDs []string) (map[string][]string, error) {
	query := fmt.Sprintf(
		findKanaBatchQuery,
		strings.Join(sqlutils.GetPlaceholders(len(wordIDs)), ","),
	)

	args := make([]any, len(wordIDs))
	for i, id := range wordIDs {
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	result := make(map[string][]string, len(wordIDs))

	for rows.Next() {
		var wordID, text string

		if err := rows.Scan(&wordID, &text); err != nil {
			return nil, err
		}

		result[wordID] = append(result[wordID], text)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *SQLiteRepo) loadTranslationsBatch(ctx context.Context, wordIDs []string) (map[string][]Translation, error) {
	placeholders := strings.Join(sqlutils.GetPlaceholders(len(wordIDs)), ",")
	query := fmt.Sprintf(findTranslationsBatchQuery, placeholders, placeholders)

	args := make([]any, 0, 2*len(wordIDs))
	for _, id := range wordIDs {
		args = append(args, id)
	}
	for _, id := range wordIDs {
		args = append(args, id)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	result := make(map[string][]Translation, len(wordIDs))

	var currentWordID string
	var currentSenseID int64 = -1

	for rows.Next() {
		var got struct {
			wordID  string
			senseID int64
			lang    string
			text    string
			pos     *string
		}

		if err := rows.Scan(&got.wordID, &got.senseID, &got.lang, &got.text, &got.pos); err != nil {
			return nil, err
		}

		translations := result[got.wordID]

		if len(translations) == 0 || got.wordID != currentWordID || got.senseID != currentSenseID {
			translations = append(translations, Translation{
				SenseID: got.senseID,
				Pos:     got.pos,
			})
			result[got.wordID] = translations
			currentWordID = got.wordID
			currentSenseID = got.senseID
		}

		last := &translations[len(translations)-1]

		last.Glosses = append(last.Glosses, Gloss{
			Lang: got.lang,
			Text: got.text,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *SQLiteRepo) loadKanaReadingsForKanjiBatch(ctx context.Context, wordIDs []string) (map[string]map[string]string, error) {
	query := fmt.Sprintf(
		findKanaReadingsForKanjiBatchQuery,
		strings.Join(sqlutils.GetPlaceholders(len(wordIDs)), ","),
	)

	args := make([]any, len(wordIDs))
	for i, id := range wordIDs {
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	result := make(map[string]map[string]string, len(wordIDs))

	for rows.Next() {
		var wordID, kanjiText, kanaText string

		if err := rows.Scan(&wordID, &kanjiText, &kanaText); err != nil {
			return nil, err
		}

		readings, ok := result[wordID]
		if !ok {
			readings = make(map[string]string)
			result[wordID] = readings
		}

		if kanjiText == "*" {
			if _, exists := readings["*"]; !exists {
				readings["*"] = kanaText
			}
		} else {
			if _, exists := readings[kanjiText]; !exists {
				readings[kanjiText] = kanaText
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, readings := range result {
		if wildcard, ok := readings["*"]; ok && wildcard != "" {
			for k := range readings {
				if readings[k] == "" {
					readings[k] = wildcard
				}
			}
		}
	}

	return result, nil
}

func buildEntryRuby(kanji []string, kana []string, kanjiReadings map[string]string) []RubySegment {
	if len(kanji) == 0 {
		return nil
	}

	firstKanji := kanji[0]

	if reading, ok := kanjiReadings[firstKanji]; ok {
		return BuildRubySegments(firstKanji, reading)
	}

	if wildcard, ok := kanjiReadings["*"]; ok {
		return BuildRubySegments(firstKanji, wildcard)
	}

	if len(kana) > 0 {
		return BuildRubySegments(firstKanji, kana[0])
	}

	return []RubySegment{{Text: firstKanji}}
}
