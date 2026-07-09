package jmdict

import (
	"database/sql"
	_ "embed"
)

//go:embed sql/schema.sql
var schema string

func InitJMDictDB(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return nil
}
