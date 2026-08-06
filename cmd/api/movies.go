package main

import (
	"net/http"
	"time"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/validator"
)

// movie calls a.ok with wrap{"movie":m} as data.
func (a *app) movie(w http.ResponseWriter, r *http.Request, m *data.Movie) {
	a.ok(w, r, wrap{"movie": m})
}

func (a *app) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := id(r)
	if err != nil {
		a.notFound(w, r)
		return
	}
	m := &data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	a.movie(w, r, m)
}

type postInput struct {
	Title   *string       `json:"title"`
	Year    *int32        `json:"year"`
	Runtime *data.Runtime `json:"runtime"`
	Genres  []string      `json:"genres"`
}

func validatePostInput(v *validator.Validator, i *postInput) {
	mustProvided := "must be provided"
	v.Validate(i.Title != nil, "title", mustProvided)
	v.Validate(i.Year != nil, "year", mustProvided)
	v.Validate(i.Runtime != nil, "runtime", mustProvided)
}

func movieFromPostInput(i *postInput) *data.Movie {
	m := &data.Movie{
		Genres: i.Genres,
	}
	if i.Title != nil {
		m.Title = *i.Title
	}
	if i.Year != nil {
		m.Year = *i.Year
	}
	if i.Runtime != nil {
		m.Runtime = *i.Runtime
	}

	return m
}

func (a *app) postMovieHandler(w http.ResponseWriter, r *http.Request) {
	i := new(postInput)
	if err := decodeJSON(w, r, i); err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	validatePostInput(v, i)

	m := movieFromPostInput(i)
	if data.ValidateMovie(v, m); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	a.movie(w, r, m)
}
