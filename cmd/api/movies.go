package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/validator"
)

// movie calls a.ok with wrap{"movie":m} as data.
func (a *app) movie(w http.ResponseWriter, r *http.Request, m *data.Movie) {
	w.Header().Set("Location", fmt.Sprintf("/v1/movies/%v", m.ID))
	a.ok(w, r, wrap{"movie": m})
}

func (a *app) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := id(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	m, err := a.repo.Movies.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFound(w, r)
		} else {
			a.serverErr(w, r, err)
		}
		return
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
	err := decodeJSON(w, r.Body, &i)
	if err != nil {
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

	if err = a.repo.Movies.Add(m); err != nil {
		a.serverErr(w, r, err)
		return
	}

	a.movie(w, r, m)
}

func (a *app) putMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := id(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	m, err := a.repo.Movies.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFound(w, r)
		} else {
			a.serverErr(w, r, err)
		}
		return
	}

	var i struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	if err = decodeJSON(w, r.Body, &i); err != nil {
		a.badRequest(w, r, err)
		return
	}

	m.Title = i.Title
	m.Year = i.Year
	m.Runtime = i.Runtime
	m.Genres = i.Genres
	v := validator.New()

	if data.ValidateMovie(v, m); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	if err = a.repo.Movies.Set(m); err != nil {
		a.serverErr(w, r, err)
		return
	}

	a.movie(w, r, m)
}

func (a *app) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := id(r)
	if err != nil {
		a.notFound(w, r)
		return
	}

	if err = a.repo.Movies.Delete(id); err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFound(w, r)
		} else {
			a.serverErr(w, r, err)
		}
		return
	}

	a.succeeded(w, r)
}
