package data

import (
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("data: record not found")

// dester represents a struct whose dest method returns pointers to each of the struct's fields.
// This is for scanning querying db result into struct fields.
type dester interface {
	dest() []any
}

type Data struct {
	Movies MovieData
}

func New(db *sql.DB) *Data {
	return &Data{
		Movies: MovieData{db},
	}
}

func queryRowAndScan(db *sql.DB, query string, args []any, dest []any) error {
	if err := db.QueryRow(query, args...).Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func rowsAffected(db *sql.DB, query string, args ...any) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
