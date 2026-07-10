// Package repository provides API for accessing JMDict Database
package repository

import (
	"database/sql"
)

type SQLiteRepo struct {
	db *sql.DB

	find *sql.Stmt

	findByKanji *sql.Stmt
	findByKana  *sql.Stmt

	findKanji *sql.Stmt
	findKana  *sql.Stmt

	findTranslations *sql.Stmt
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

	return repo, nil
}

func (r *SQLiteRepo) FindByKanji(text string) ([]Entry, error) {
	rows, err := r.findByKanji.Query(text)
	if err != nil {
		return nil, err
	}

	return r.loadEntries(rows)
}

func (r *SQLiteRepo) FindByKana(text string) ([]Entry, error) {
	rows, err := r.findByKana.Query(text)
	if err != nil {
		return nil, err
	}

	return r.loadEntries(rows)
}

func (r *SQLiteRepo) Find(text string) ([]Entry, error) {
	rows, err := r.find.Query(text, text)
	if err != nil {
		return nil, err
	}

	return r.loadEntries(rows)
}

func (r *SQLiteRepo) loadEntries(rows *sql.Rows) ([]Entry, error) {
	defer func() {
		_ = rows.Close()
	}()

	var entries []Entry

	for rows.Next() {
		var wordID string

		if err := rows.Scan(&wordID); err != nil {
			return nil, err
		}

		entry, err := r.loadEntry(wordID)
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

func (r *SQLiteRepo) loadKanji(wordID string) ([]string, error) {
	rows, err := r.findKanji.Query(wordID)
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

func (r *SQLiteRepo) loadKana(wordID string) ([]string, error) {
	rows, err := r.findKana.Query(wordID)
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

func (r *SQLiteRepo) loadTranslations(wordID string) ([]Translation, error) {
	rows, err := r.findTranslations.Query(wordID)
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

func (r *SQLiteRepo) loadEntry(wordID string) (Entry, error) {
	kanji, err := r.loadKanji(wordID)
	if err != nil {
		return Entry{}, err
	}

	kana, err := r.loadKana(wordID)
	if err != nil {
		return Entry{}, err
	}

	translations, err := r.loadTranslations(wordID)
	if err != nil {
		return Entry{}, err
	}

	return Entry{
		ID:           wordID,
		Kanji:        kanji,
		Kana:         kana,
		Translations: translations,
	}, nil
}
