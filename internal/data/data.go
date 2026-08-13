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

// dester represents a pointer receiver of a struct, whose dest method returns a slice of pointers to each of the struct's fields.
// This is for scanning db query result into struct fields.
type dester interface {
	dest() []any
}

type Repo struct {
	Movies MovieRepo
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		Movies: MovieRepo{db},
	}
}

// queryRow queries a row from db and scan each row's column to each of dest.
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
// rowsAffected times out after timeoutDuration (const).
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
// query times out after timeoutDuration (const).
func query[T dester](db *sql.DB, query string, args []any, newDester func() T) ([]T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ts := []T{}
	for rows.Next() {
		t := newDester()
		if err = rows.Scan(t.dest()...); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ts, nil
}
