package data

import (
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("data: record not found")

type Data struct {
	Movies MovieData
}

func New(db *sql.DB) *Data {
	return &Data{
		Movies: MovieData{db},
	}
}
