package data

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/tanvmn/mdb/internal/validator"
)

const (
	MinMovieYear = 1888

	selectMovies      = "select m.id, m.created_at, m.title, m.year, m.runtime, m.genres, m.version from movies as m "
	updateMovies      = "update movies set title=$1, year=$2, runtime=$3, genres=$4, version = version+1 "
	updateMoviesWhere = updateMovies + " where "
	returningMovie    = " returning id, created_at, title, year, runtime, genres, version "
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero"` // don't embed Runtime, cause (*)Movie will satisfies json.Marshaler, then JSON of (*)Movie will be JSON of (*)Runtime.
	Genres    []string  `json:"genres,omitzero"`
	Version   int32     `json:"version"`
}

func (m *Movie) dest() []any {
	return []any{
		&m.ID,
		&m.CreatedAt,
		&m.Title,
		&m.Year,
		&m.Runtime,
		pq.Array(&m.Genres),
		&m.Version,
	}
}

func ValidateMovie(v *validator.Validator, m *Movie) {
	v.Validate(m.Title != "", "title", "must be provided and not an empty string")
	v.Validate(len(m.Title) <= 500, "title", "must not be longer than 500 bytes")

	v.Validate(m.Year != 0, "year", fmt.Sprintf("must be provided and not before %v", MinMovieYear))
	v.Validate(m.Year >= MinMovieYear, "year", fmt.Sprintf("must not be before %v", MinMovieYear))
	v.Validate(m.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Validate(m.Runtime != 0, "runtime", "must be provided and a positive integer")
	v.Validate(m.Runtime > 0, "runtime", "must be a positive integer")

	genresCount := len(m.Genres)
	v.Validate(m.Genres != nil, "genres", "must be provided")
	v.Validate(genresCount >= 1, "genres", "must have at least 1")
	v.Validate(genresCount <= 5, "genres", "must not have more than 5")
	v.Validate(validator.Unique(m.Genres), "genres", "must not have duplicates")
}

type MovieData struct {
	db *sql.DB
}

func (d *MovieData) Add(m *Movie) error {
	q := `insert into movies (title, year, runtime, genres) values ($1,$2,$3,$4)
	returning id, created_at, title, year, runtime, genres, version
	`
	// pq.Array adapts a Go value to a value that postgres understands.
	args := []any{m.Title, m.Year, m.Runtime, pq.Array(m.Genres)}

	return queryRowAndScan(d.db, q, args, m.dest())
}

func (d *MovieData) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	q := selectMovies + " where id=$1"
	args := []any{id}
	m := new(Movie)
	err := queryRowAndScan(d.db, q, args, m.dest())

	return m, err
}

func (d *MovieData) Set(m *Movie) error {
	q := updateMoviesWhere + " id=$5 " + returningMovie
	args := []any{m.Title, m.Year, m.Runtime, pq.Array(m.Genres), m.ID}

	return queryRowAndScan(d.db, q, args, m.dest())
}

func (d *MovieData) Delete(id int64) error {
	q := "delete from movies where id=$1"
	ra, err := rowsAffected(d.db, q, id)
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrRecordNotFound
	}

	return nil
}
