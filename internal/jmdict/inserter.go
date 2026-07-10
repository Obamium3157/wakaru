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

func (ins *inserter) exec(stmt *sql.Stmt, args ...any) error {
	_, err := stmt.Exec(args...)
	return err
}

func (ins *inserter) execInsertID(stmt *sql.Stmt, args ...any) (int64, error) {
	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (ins *inserter) tagID(tagValue tag) (int64, error) {
	id, ok := ins.tagIDs[tagValue]
	if !ok {
		return 0, fmt.Errorf("unknown tag %q", tagValue)
	}

	return id, nil
}

func (ins *inserter) insertWordFn(word *Word) error {
	if err := ins.exec(ins.insertWord, word.ID); err != nil {
		return err
	}

	for order, kanaValue := range word.Kana {
		if err := ins.insertKanaFn(word.ID, &kanaValue, order); err != nil {
			return err
		}
	}

	for order, kanjiValue := range word.Kanji {
		if err := ins.insertKanjiFn(word.ID, &kanjiValue, order); err != nil {
			return err
		}
	}

	for order, senseValue := range word.Senses {
		if err := ins.insertSenseFn(word.ID, &senseValue, order); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) insertKanaFn(wordID string, kanaValue *kana, order int) error {
	kanaID, err := ins.execInsertID(
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
		if err := ins.exec(
			ins.insertKanaAppliesToKanji,
			kanaID,
			kanjiText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for _, tagValue := range kanaValue.Tags {
		tagID, err := ins.tagID(tagValue)
		if err != nil {
			return err
		}

		if err := ins.exec(
			ins.insertKanaTag,
			kanaID,
			tagID,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) insertKanjiFn(wordID string, kanjiValue *kanji, order int) error {
	kanjiID, err := ins.execInsertID(
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
		tagID, err := ins.tagID(tagValue)
		if err != nil {
			return err
		}

		if err := ins.exec(
			ins.insertKanjiTag,
			kanjiID,
			tagID,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ins *inserter) insertSenseFn(wordID string, senseValue *Sense, order int) error {
	senseID, err := ins.execInsertID(
		ins.insertSense,
		wordID,
		order,
	)
	if err != nil {
		return err
	}

	for ordNum, pos := range senseValue.PartOfSpeech {
		tagID, err := ins.tagID(pos)
		if err != nil {
			return err
		}

		if err := ins.exec(ins.insertPartOfSpeech, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, field := range senseValue.Field {
		tagID, err := ins.tagID(field)
		if err != nil {
			return err
		}

		if err := ins.exec(ins.insertField, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, dialect := range senseValue.Dialect {
		tagID, err := ins.tagID(dialect)
		if err != nil {
			return err
		}

		if err := ins.exec(ins.insertDialect, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, misc := range senseValue.Misc {
		tagID, err := ins.tagID(misc)
		if err != nil {
			return err
		}

		if err := ins.exec(ins.insertMisc, senseID, tagID, ordNum); err != nil {
			return err
		}
	}

	for ordNum, kanaText := range senseValue.AppliesToKana {
		if err := ins.exec(
			ins.insertSenseAppliesToKana,
			senseID,
			kanaText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, kanjiText := range senseValue.AppliesToKanji {
		if err := ins.exec(
			ins.insertSenseAppliesToKanji,
			senseID,
			kanjiText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, infoText := range senseValue.Info {
		if err := ins.exec(
			ins.insertInfo,
			senseID,
			infoText,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, source := range senseValue.LanguageSource {
		if err := ins.exec(
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
		if err := ins.exec(
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
		if err := ins.insertXrefRow(
			senseID,
			"related",
			related,
			ordNum,
		); err != nil {
			return err
		}
	}

	for ordNum, antonym := range senseValue.Antonym {
		if err := ins.insertXrefRow(
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

func (ins *inserter) insertXrefRow(
	senseID int64,
	relationType string,
	ref xref,
	order int,
) error {
	return ins.exec(
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
