package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("data: record not found")
	ErrEditConflict   = errors.New("data: edit conflict")
)

// dester represents a struct whose dest method returns pointers to each of the struct's fields.
// This is for scanning querying db result into struct fields.
type dester interface {
	dest() []any
}

type Repo struct {
	Movies MovieRepo
}

func New(db *sql.DB) *Repo {
	return &Repo{
		Movies: MovieRepo{db},
	}
}

// queryRowAndScan queries a row from db and scan each column to each of dest.
// ErrRecordNotFound is returned if db returns no rows.
func queryRowAndScan(db *sql.DB, query string, args []any, dest []any) error {
	if err := db.QueryRow(query, args...).Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

// rowsAffected executes query in db and returned the rows affected.
func rowsAffected(db *sql.DB, query string, args ...any) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
