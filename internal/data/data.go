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
	// ErrEditConflict will be returned when race condition (usually data race) occurs while multiple go routines update the same record.
	ErrEditConflict = errors.New("data: edit conflict")
)

// dester represents a pointer to a struct.
// dest() returns a slice of pointers to each of the struct's fields.
// This is for scanning db query result into struct fields.
type dester interface {
	dest() []any
}

type Repo struct {
	Movies MovieRepo
	Tokens TokenRepo
	Users  UserRepo
	Perms  PermRepo
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		Movies: MovieRepo{db},
		Tokens: TokenRepo{db},
		Users:  UserRepo{db},
		Perms:  PermRepo{db},
	}
}

// queryRow queries a row and scan each row's column to each of dest.
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

// result executes query and returns the result.
// result times out after timeoutDuration (const).
func result(db *sql.DB, query string, args []any) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	return db.ExecContext(ctx, query, args...)
}

// rowsAffected executes query and returned the rows affected.
func rowsAffected(db *sql.DB, query string, args []any) (int64, error) {
	result, err := result(db, query, args)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// query queries rows and returns them as []T.
// T's argument must be a pointer to struct to enable mutating.
// newT must return a non-nil pointer.
//
//	Ex: if T is *Movie, then newT is func() *Movie, and newT() != nil.
//
// total is the number of rows before pagination.
// query times out after timeoutDuration (const).
func query[T dester](db *sql.DB, query string, args []any, newT func() T) (ts []T, total int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err2 := rows.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	ts = []T{}
	for rows.Next() {
		t := newT()
		dest := append([]any{&total}, t.dest()...)
		if err = rows.Scan(dest...); err != nil {
			return nil, 0, err
		}
		ts = append(ts, t)
	}

	return ts, total, rows.Err()
}
