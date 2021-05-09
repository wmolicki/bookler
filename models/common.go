package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// first will scan the row into dest object, dest must be pointer type
func first(dest interface{}, row *sqlx.Row) error {
	err := row.StructScan(dest)
	if err == sql.ErrNoRows {
		return ErrorEntityNotFound
	}
	return err
}
