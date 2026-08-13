// Package data contains definitions and logic about database and its data.
package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const timeoutDuration = 3 * time.Second

var (
	ErrRecordNotFound = errors.New("data: record not found")
	ErrEditConflict   = errors.New("data: edit conflict")
)

// queryDester represents a struct whose queryDest method returns a pointer to the struct and a slice of pointers to each of the struct's fields.
// This is for scanning db query result into struct fields.
type queryDester interface {
	queryDest() (queryDester, []any)
}

type Repo struct {
	Movies MovieRepo
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		Movies: MovieRepo{db},
	}
}

// queryRow queries a row from db and scan each column to each of dest.
// queryRow times out after timeoutDuration (const).
// ErrRecordNotFound is returned when db returns no rows.
func queryRow(db *sql.DB, query string, args []any, dest []any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	if err := db.QueryRowContext(ctx, query, args...).Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

// rowsAffected executes query in db and returned the rows affected.
// rowsAffected will time out if db doesn't response after timeoutDuration (const).
func rowsAffected(db *sql.DB, query string, args ...any) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// query queries rows from db and returns them as []T.
// query will time out if db doesn't response after timeoutDuration (const).
func query[T queryDester](db *sql.DB, query string, args []any) ([]T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ts := []T{}
	for rows.Next() {
		var t T
		qd, dest := t.queryDest()
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}
		t, ok := qd.(T)
		if !ok {
			return nil, errors.New("data: queryRows: concrete type of interface queryDest value, which is returned from T.queryDest(), isn't T")
		}
		ts = append(ts, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ts, nil
}
