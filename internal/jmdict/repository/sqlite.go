// Package repository provides API for accessing JMDict Database
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"wakaru/internal/sqlutils"
)

type SQLiteRepo struct {
	db *sql.DB

	find *sql.Stmt

	findByKanji *sql.Stmt
	findByKana  *sql.Stmt

	findKanji *sql.Stmt
	findKana  *sql.Stmt

	findTranslations *sql.Stmt

	findAllForms *sql.Stmt

	findKanaReadingsForKanji *sql.Stmt
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
	if repo.findKanji, err = db.Prepare(findKanjiQuery); err != nil {
		return nil, err
	}
	if repo.findKana, err = db.Prepare(findKanaQuery); err != nil {
		return nil, err
	}
	if repo.findTranslations, err = db.Prepare(findTranslationsQuery); err != nil {
		return nil, err
	}
	if repo.findAllForms, err = db.Prepare(findAllFormsQuery); err != nil {
		return nil, err
	}
	if repo.findKanaReadingsForKanji, err = db.Prepare(findKanaReadingsForKanjiQuery); err != nil {
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
		_ = rows.Close()
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

func (r *SQLiteRepo) loadEntries(ctx context.Context, rows *sql.Rows) ([]Entry, error) {
	defer func() {
		_ = rows.Close()
	}()

	var entries []Entry

	for rows.Next() {
		var wordID string

		if err := rows.Scan(&wordID); err != nil {
			return nil, err
		}

		entry, err := r.loadEntry(ctx, wordID)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (r *SQLiteRepo) loadKanji(ctx context.Context, wordID string) ([]string, error) {
	rows, err := r.findKanji.QueryContext(ctx, wordID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var result []string

	for rows.Next() {
		var text string

		if err := rows.Scan(&text); err != nil {
			return nil, err
		}

		result = append(result, text)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *SQLiteRepo) loadKana(ctx context.Context, wordID string) ([]string, error) {
	rows, err := r.findKana.QueryContext(ctx, wordID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var result []string

	for rows.Next() {
		var text string

		if err := rows.Scan(&text); err != nil {
			return nil, err
		}

		result = append(result, text)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *SQLiteRepo) loadTranslations(ctx context.Context, wordID string) ([]Translation, error) {
	rows, err := r.findTranslations.QueryContext(ctx, wordID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var translations []Translation

	var currentSenseID int64 = -1

	for rows.Next() {
		var got struct {
			senseID int64
			lang    string
			text    string
		}

		if err := rows.Scan(&got.senseID, &got.lang, &got.text); err != nil {
			return nil, err
		}

		if len(translations) == 0 || got.senseID != currentSenseID {
			translations = append(translations, Translation{
				SenseID: got.senseID,
			})
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

	return translations, nil
}

func (r *SQLiteRepo) loadEntry(ctx context.Context, wordID string) (Entry, error) {
	kanji, err := r.loadKanji(ctx, wordID)
	if err != nil {
		return Entry{}, err
	}

	kana, err := r.loadKana(ctx, wordID)
	if err != nil {
		return Entry{}, err
	}

	translations, err := r.loadTranslations(ctx, wordID)
	if err != nil {
		return Entry{}, err
	}

	kanjiReadings, err := r.loadKanaReadingsForKanji(ctx, wordID)
	if err != nil {
		return Entry{}, err
	}

	ruby := buildEntryRuby(kanji, kana, kanjiReadings)

	return Entry{
		ID:           wordID,
		Kanji:        kanji,
		Kana:         kana,
		Ruby:         ruby,
		Translations: translations,
	}, nil
}

func (r *SQLiteRepo) loadKanaReadingsForKanji(ctx context.Context, wordID string) (map[string]string, error) {
	rows, err := r.findKanaReadingsForKanji.QueryContext(ctx, wordID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make(map[string]string)
	var wildcardReading string

	for rows.Next() {
		var kanjiText, kanaText string
		if err := rows.Scan(&kanjiText, &kanaText); err != nil {
			return nil, err
		}
		if kanjiText == "*" {
			if wildcardReading == "" {
				wildcardReading = kanaText
			}
		} else {
			if _, exists := result[kanjiText]; !exists {
				result[kanjiText] = kanaText
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if wildcardReading != "" {
		for k := range result {
			if result[k] == "" {
				result[k] = wildcardReading
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
