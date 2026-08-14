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
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	if err := decodeJSON(w, r.Body, &i); err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	v.Validate(i.Title != nil, "title", "must be provided")
	v.Validate(i.Year != nil, "year", "must be provided")
	v.Validate(i.Runtime != nil, "runtime", "must be provided")

	m := &data.Movie{Genres: i.Genres}
	if i.Title != nil {
		m.Title = *i.Title
	}
	if i.Year != nil {
		m.Year = *i.Year
	}
	if i.Runtime != nil {
		m.Runtime = *i.Runtime
	}
	if data.ValidateMovie(v, m); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	if err := a.repo.Movies.Add(m); err != nil {
		a.serverErr(w, r, err)
		return
	}

	a.movie(w, r, m)
}

func (a *app) patchMovieHandler(w http.ResponseWriter, r *http.Request) {
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
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	if err = decodeJSON(w, r.Body, &i); err != nil {
		a.badRequest(w, r, err)
		return
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
	if i.Genres != nil {
		m.Genres = i.Genres
	}

	v := validator.New()
	if data.ValidateMovie(v, m); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	// retrying a few times doesn't seem to lessen failure rate.
	if err = a.repo.Movies.Set(m); err != nil {
		if errors.Is(err, data.ErrEditConflict) {
			a.editConflict(w, r)
		} else {
			a.serverErr(w, r, err)
		}
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

// movies calls a.ok with wrap{"movies":ms} as data.
func (a *app) movies(w http.ResponseWriter, r *http.Request, ms []*data.Movie) {
	a.ok(w, r, wrap{"movies": ms})
}

func (a *app) getMoviesHandler(w http.ResponseWriter, r *http.Request) {
	v := validator.New()
	q := r.URL.Query()
	var i struct {
		Title  string
		Genres []string
		data.Filters
	}
	i.Title = parseString(q, "title", "")
	i.Genres = parseCSV(q, "genres", []string{})
	i.Page = parseInt(q, "page", 1, v)
	i.PageSize = parseInt(q, "page_size", 20, v)
	i.Sort = parseString(q, "sort", "id")
	i.SortList = []string{"id", "-id", "title", "-title", "year", "-year", "runtime", "-runtime"}
	// at this point, service could serve 400 if !v.Valid(), but the next step is still using v to validate input,
	// so continue validating and serve 422 if !v.Valid().

	if data.ValidateFilters(v, &i.Filters); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	ms, err := a.repo.Movies.GetMany(i.Title, i.Genres, &i.Filters)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	a.movies(w, r, ms)
}
