package main

import (
	"net/http"
	"time"

	"github.com/tanvmn/mdb/internal/data"
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
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		a.badRequest(w, r, err)
		return
	}
	a.ok(w, r, input)
}
