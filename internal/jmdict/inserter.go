package jmdict

import (
	"database/sql"
	"fmt"
	"strings"
)

type inserter struct {
	tx *sql.Tx

	tagIDs map[tag]int64

	insertWord *sql.Stmt

	insertKana               *sql.Stmt
	insertKanaAppliesToKanji *sql.Stmt
	insertKanaTag            *sql.Stmt

	insertKanji    *sql.Stmt
	insertKanjiTag *sql.Stmt

	insertSense               *sql.Stmt
	insertPartOfSpeech        *sql.Stmt
	insertSenseAppliesToKanji *sql.Stmt
	insertSenseAppliesToKana  *sql.Stmt
	insertXref                *sql.Stmt
	insertField               *sql.Stmt
	insertDialect             *sql.Stmt
	insertMisc                *sql.Stmt
	insertInfo                *sql.Stmt
	insertLanguageSource      *sql.Stmt
	insertGloss               *sql.Stmt

	prepared []*sql.Stmt
}

func newInserter(tx *sql.Tx, dict *Dictionary) (*inserter, error) {
	ins := &inserter{
		tx:     tx,
		tagIDs: make(map[tag]int64),
	}

	if err := ins.loadTags(dict); err != nil {
		return nil, err
	}

	if err := ins.initStmts(); err != nil {
		ins.Close()
		return nil, err
	}

	return ins, nil
}

func (ins *inserter) Exec(stmt *sql.Stmt, args ...any) error {
	_, err := stmt.Exec(args...)
	return err
}

func (ins *inserter) ExecInsertID(stmt *sql.Stmt, args ...any) (int64, error) {
	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (ins *inserter) TagID(tagValue tag) (int64, error) {
	id, ok := ins.tagIDs[tagValue]
	if !ok {
		return 0, fmt.Errorf("unknown tag %q", tagValue)
	}

	return id, nil
}

func (ins *inserter) InsertWord(word *Word) error {
	if err := ins.Exec(ins.insertWord, word.ID); err != nil {
		return err
	}

	for order, kanaValue := range word.Kana {
		if err := ins.InsertKana(word.ID, &kanaValue, order); err != nil {
			return err
		}
	}

	for order, kanjiValue := range word.Kanji {
		if err := ins.InsertKanji(word.ID, &kanjiValue, order); err != nil {
			return err
		}
	}

	for order, senseValue := range word.Senses {
		if err := ins.InsertSense(word.ID, &senseValue, order); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) InsertKana(wordID string, kanaValue *kana, order int) error {
	kanaID, err := ins.ExecInsertID(
		ins.insertKana,
		wordID,
		kanaValue.Common,
		kanaValue.Text,
		order,
	)
	if err != nil {
		return err
	}

	for ordNum, kanjiText := range kanaValue.AppliesToKanji {
		if err := ins.Exec(
			ins.insertKanaAppliesToKanji,
			kanaID,
			kanjiText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for _, tagValue := range kanaValue.Tags {
		tagID, err := ins.TagID(tagValue)
		if err != nil {
			return err
		}

		if err := ins.Exec(
			ins.insertKanaTag,
			kanaID,
			tagID,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) InsertKanji(wordID string, kanjiValue *kanji, order int) error {
	kanjiID, err := ins.ExecInsertID(
		ins.insertKanji,
		wordID,
		kanjiValue.Common,
		kanjiValue.Text,
		order,
	)
	if err != nil {
		return err
	}

	for _, tagValue := range kanjiValue.Tags {
		tagID, err := ins.TagID(tagValue)
		if err != nil {
			return err
		}

		if err := ins.Exec(
			ins.insertKanjiTag,
			kanjiID,
			tagID,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) InsertSense(wordID string, senseValue *Sense, order int) error {
	senseID, err := ins.ExecInsertID(
		ins.insertSense,
		wordID,
		order,
	)
	if err != nil {
		return err
	}

	for ordNum, pos := range senseValue.PartOfSpeech {
		tagID, err := ins.TagID(pos)
		if err != nil {
			return err
		}

		if err := ins.Exec(ins.insertPartOfSpeech, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, field := range senseValue.Field {
		tagID, err := ins.TagID(field)
		if err != nil {
			return err
		}

		if err := ins.Exec(ins.insertField, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, dialect := range senseValue.Dialect {
		tagID, err := ins.TagID(dialect)
		if err != nil {
			return err
		}

		if err := ins.Exec(ins.insertDialect, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, misc := range senseValue.Misc {
		tagID, err := ins.TagID(misc)
		if err != nil {
			return err
		}

		if err := ins.Exec(ins.insertMisc, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, kanaText := range senseValue.AppliesToKana {
		if err := ins.Exec(
			ins.insertSenseAppliesToKana,
			senseID,
			kanaText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, kanjiText := range senseValue.AppliesToKanji {
		if err := ins.Exec(
			ins.insertSenseAppliesToKanji,
			senseID,
			kanjiText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, infoText := range senseValue.Info {
		if err := ins.Exec(
			ins.insertInfo,
			senseID,
			infoText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, source := range senseValue.LanguageSource {
		if err := ins.Exec(
			ins.insertLanguageSource,
			senseID,
			source.Lang,
			source.Text,
			source.Full,
			source.Wasei,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, glossValue := range senseValue.Gloss {
		if err := ins.Exec(
			ins.insertGloss,
			senseID,
			glossValue.Lang,
			glossValue.Text,
			glossValue.Type,
			glossValue.Gender,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, related := range senseValue.Related {
		if err := ins.InsertXrefRow(
			senseID,
			"related",
			related,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, antonym := range senseValue.Antonym {
		if err := ins.InsertXrefRow(
			senseID,
			"antonym",
			antonym,
			ordNum,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) InsertXrefRow(
	senseID int64,
	relationType string,
	ref xref,
	order int,
) error {
	return ins.Exec(
		ins.insertXref,
		senseID,
		relationType,
		ref.Headword,
		ref.Reading,
		ref.SenseIndex,
		order,
	)
}

func (ins *inserter) initStmts() error {
	senseTagColumns := []string{
		"sense_id",
		"tag_id",
		"display_order",
	}

	kanaKanjiColumns := []string{
		"word_id",
		"is_common",
		"text",
		"display_order",
	}

	statements := []struct {
		dest    **sql.Stmt
		table   string
		columns []string
	}{
		{
			&ins.insertWord,
			"word",
			[]string{
				"id",
			},
		},
		{
			&ins.insertKana,
			"kana",
			kanaKanjiColumns,
		},
		{
			&ins.insertKanaAppliesToKanji,
			"kana_applies_to_kanji",
			[]string{
				"kana_id",
				"kanji_text",
				"display_order",
			},
		},
		{
			&ins.insertKanaTag,
			"kana_tag",
			[]string{
				"kana_id",
				"tag_id",
			},
		},
		{
			&ins.insertKanji,
			"kanji",
			kanaKanjiColumns,
		},
		{
			&ins.insertKanjiTag,
			"kanji_tag",
			[]string{
				"kanji_id",
				"tag_id",
			},
		},
		{
			&ins.insertSense,
			"sense",
			[]string{
				"word_id",
				"display_order",
			},
		},
		{
			&ins.insertPartOfSpeech,
			"sense_part_of_speech",
			senseTagColumns,
		},
		{
			&ins.insertField,
			"sense_field",
			senseTagColumns,
		},
		{
			&ins.insertDialect,
			"sense_dialect",
			senseTagColumns,
		},
		{
			&ins.insertMisc,
			"sense_misc",
			senseTagColumns,
		},
		{
			&ins.insertSenseAppliesToKana,
			"sense_applies_to_kana",
			[]string{
				"sense_id",
				"kana_text",
				"display_order",
			},
		},
		{
			&ins.insertSenseAppliesToKanji,
			"sense_applies_to_kanji",
			[]string{
				"sense_id",
				"kanji_text",
				"display_order",
			},
		},
		{
			&ins.insertInfo,
			"sense_info",
			[]string{
				"sense_id",
				"text",
				"display_order",
			},
		},
		{
			&ins.insertLanguageSource,
			"language_source",
			[]string{
				"sense_id",
				"lang",
				"text",
				"is_full",
				"is_wasei",
				"display_order",
			},
		},
		{
			&ins.insertGloss,
			"gloss",
			[]string{
				"sense_id",
				"lang",
				"text",
				"gloss_type",
				"gender",
				"display_order",
			},
		},
		{
			&ins.insertXref,
			"xref",
			[]string{
				"sense_id",
				"relation_type",
				"headword",
				"reading",
				"sense_index",
				"display_order",
			},
		},
	}

	for _, stmt := range statements {
		if err := ins.prepare(stmt.dest, stmt.table, stmt.columns...); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) prepare(dest **sql.Stmt, table string, columns ...string) error {
	stmt, err := ins.tx.Prepare(
		formInsertQuery(table, columns),
	)
	if err != nil {
		return err
	}

	*dest = stmt
	ins.prepared = append(ins.prepared, stmt)

	return nil
}

func (ins *inserter) Close() {
	for i, stmt := range ins.prepared {
		if stmt != nil {
			_ = stmt.Close()
			ins.prepared[i] = nil
		}
	}
}

func (ins *inserter) loadTags(dict *Dictionary) error {
	stmt, err := ins.tx.Prepare(
		`INSERT INTO tag (label, description) VALUES (?, ?)`,
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()

	for label, desc := range dict.Tags {
		res, err := stmt.Exec(label, desc)
		if err != nil {
			return err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return err
		}

		ins.tagIDs[label] = id
	}

	return nil
}

func formInsertQuery(table string, columns []string) string {
	placeholders := getPlaceholders(len(columns))

	return fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
}

func getPlaceholders(n int) []string {
	placeholders := make([]string, n)
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return placeholders
}
