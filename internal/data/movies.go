package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/tanvmn/mdb/internal/validator"
)

const (
	MinMovieYear = 1888

	selectMovies         = "select m.id, m.created_at, m.title, m.year, m.runtime, m.genres, m.version from movies as m "
	selectCountAndMovies = "select count(*) over(), m.id, m.created_at, m.title, m.year, m.runtime, m.genres, m.version from movies as m "
	// updateMoviesWhereVersion = "update movies set title=$1, year=$2, runtime=$3, genres=$4, version=version+1 where version=$5 "
	returningMovie = " returning id, created_at, title, year, runtime, genres, version "

	// direction to maybe follow: a where clause which doesn't need to be built dynamically
	// by making bool expressions of each input to either be evaluated or if not provided, be evaluated true.
	//
	// From user perspective: find me movies whose titles contain everything in title part of my input.
	// to_tsvector splits its input into lexemes.
	// plainto_tsquery converts its input to a format that postgres understands.
	// In this case, plaito_tsquery('simple', $1), which is set at 'simple' config:
	// 	normarlizes the value of $1 (don't know what this means),
	// 	removes special characters, and adds '&' operator between words. The result of plainto_tsquery('simple', '//\\@!^hello+world   ') is 'hello' & 'world'.
	// The @@ operator is the match operator.
	// The expression essentially returns true if everthing returned from plainto_tsquery appears in the lexemes returned from to_tsvector.
	//
	// Note that in 'english' config, if the input for query only contains common (stop) words, the text search expression will be ignored and evaluated to false.
	//	mdb=> select to_tsvector('simple',title) @@ plainto_tsquery('english','the') from movies;
	//	NOTICE:  text-search query contains only stop words or doesn't contain lexemes, ignored
	//	 ?column?
	//	----------
	//	 f
	//	 f
	//	 f
	filterMovies = " where ((to_tsvector('simple',title) @@ plainto_tsquery('simple',$1)) or $1='') and (genres @> $2 or $2='{}') "
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
type MovieFilter struct {
	Title  string
	Genres []string
	Filter
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
	v.Validate(m.Title != "", "title", "must not be an empty string")
	v.Validate(len(m.Title) <= 500, "title", "must not be longer than 500 bytes")

	v.Validate(m.Year >= MinMovieYear, "year", fmt.Sprintf("must not be before %v", MinMovieYear))
	v.Validate(m.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Validate(m.Runtime > 0, "runtime", "must have a positive integer")

	genresCount := len(m.Genres)
	v.Validate(m.Genres != nil, "genres", "must be provided")
	v.Validate(genresCount >= 1, "genres", "must have at least 1")
	v.Validate(genresCount <= 5, "genres", "must not have more than 5")
	v.Validate(validator.Unique(m.Genres), "genres", "must not have duplicates")
}

type MovieRepo struct {
	db *sql.DB
}

// queryMovie calls queryRow with m.dest() assigned to dest.
// ErrRecordNotFound is returned when no movie is found.
func (r *MovieRepo) queryMovie(query string, args []any, m *Movie) error {
	return queryRow(r.db, query, args, m.dest())
}

func (r *MovieRepo) Add(m *Movie) error {
	q := "insert into movies (title, year, runtime, genres) values ($1,$2,$3,$4) " + returningMovie
	// pq.Array adapts a Go value to a value that postgres understands.
	args := []any{m.Title, m.Year, m.Runtime, pq.Array(m.Genres)}

	return r.queryMovie(q, args, m)
}

func (r *MovieRepo) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	q := selectMovies + " where id=$1"
	args := []any{id}
	m := &Movie{}
	err := r.queryMovie(q, args, m)

	return m, err
}

func (r *MovieRepo) Set(m *Movie) error {
	q := "update movies set title=$1, year=$2, runtime=$3, genres=$4, version=version+1 where version=$5 and id=$6 " + returningMovie
	args := []any{m.Title, m.Year, m.Runtime, pq.Array(m.Genres), m.Version, m.ID}

	if err := r.queryMovie(q, args, m); err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrEditConflict
		}
		return err
	}

	return nil
}

func (r *MovieRepo) Delete(id int64) error {
	q := "delete from movies where id=$1"
	args := []any{id}
	ra, err := rowsAffected(r.db, q, args)
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetMany returned movies filtered and paginated by f.
func (r *MovieRepo) GetMany(f *MovieFilter) ([]*Movie, Metadata, error) {
	q := selectCountAndMovies + filterMovies + f.orderBy() + " limit $3 offset $4 "
	args := []any{f.Title, pq.Array(f.Genres), f.limit(), f.offset()}
	newMovie := func() *Movie { return &Movie{} }

	ms, total, err := query(r.db, q, args, newMovie)
	if err != nil {
		return []*Movie{}, Metadata{}, err
	}

	return ms, metadata(total, f.Page, f.PageSize), nil
}
