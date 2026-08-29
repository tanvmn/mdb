package main

import (
	"net/http"

	"github.com/tanvmn/mdb/internal/data"
)

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/health", a.getHealthHandler)

	mux.HandleFunc("GET /v1/movies", a.authorize(data.PermMoviesRead, a.getMoviesHandler))
	mux.HandleFunc("POST /v1/movies", a.authorize(data.PermMoviesWrite, a.postMovieHandler))
	mux.HandleFunc("GET /v1/movies/{id}", a.authorize(data.PermMoviesRead, a.getMovieHandler))
	mux.HandleFunc("PATCH /v1/movies/{id}", a.authorize(data.PermMoviesWrite, a.patchMovieHandler))
	mux.HandleFunc("DELETE /v1/movies/{id}", a.authorize(data.PermMoviesWrite, a.deleteMovieHandler))

	mux.HandleFunc("POST /v1/users", a.postUserHandler)
	mux.HandleFunc("PUT /v1/users/activated", a.putUserActivatedHandler)

	mux.HandleFunc("POST /v1/tokens/authentication", a.postAuthenTokenHandler)

	return chain([]middleware{a.recoverPanic, a.rateLimit, a.authenticate, a.logRequest}, mux)
}
