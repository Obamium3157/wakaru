package jmdict

import (
	"database/sql"
	"log"
)

// FillDatabase fills JMDict database with all data from given Dictionary instance
func FillDatabase(db *sql.DB, dict *Dictionary) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Printf("failed to rollback transaction: %v", err)
			}
		}
	}()

	ins, err := newInserter(tx, dict)
	if err != nil {
		return err
	}

	for _, word := range dict.Words {
		if err := ins.InsertWord(&word); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}
