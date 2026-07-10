package jmdict

import (
	"database/sql"
)

// FillDatabase fills JMDict database with all data from given Dictionary instance
func FillDatabase(db *sql.DB, dict *Dictionary) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	ins, err := newInserter(tx, dict)
	if err != nil {
		return err
	}

	for _, word := range dict.Words {
		if err := ins.insertWordFn(&word); err != nil {
			return err
		}
	}

	return tx.Commit()
}
