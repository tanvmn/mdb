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

// newDesterer's concrete type is the same as dester's.
// newDesterer's newDester returns a new instance of dester's concrete type.
type newDesterer interface {
	newDester() dester
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

func query[T newDesterer](db *sql.DB, query string, args []any) ([]T, error) {
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
		d := t.newDester()
		t, ok := d.(T)
		if !ok {
			return nil, errors.New("data: query: T.newDester() didn't return a value of T")
		}
		if err = rows.Scan(d.dest()...); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ts, nil
}
