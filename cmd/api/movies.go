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

func (a *app) postMovieHandler(w http.ResponseWriter, r *http.Request) {
	var i struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	if err := decodeJSON(w, r, &i); err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	m := &data.Movie{
		Title:   i.Title,
		Year:    i.Year,
		Runtime: i.Runtime,
		Genres:  i.Genres,
	}
	if data.ValidateMovie(v, m); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	a.movie(w, r, m)
}
