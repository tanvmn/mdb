package main

import "net/http"

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.notFound)

	mux.HandleFunc("GET /v1/health", a.getHealthHandler)

	mux.HandleFunc("GET /v1/movies", a.getMoviesHandler)
	mux.HandleFunc("POST /v1/movies", a.postMovieHandler)
	mux.HandleFunc("GET /v1/movies/{id}", a.getMovieHandler)
	mux.HandleFunc("PATCH /v1/movies/{id}", a.patchMovieHandler)
	mux.HandleFunc("DELETE /v1/movies/{id}", a.deleteMovieHandler)

	mux.HandleFunc("POST /v1/users", a.postUserHandler)
	mux.HandleFunc("PUT /v1/users/activated", a.putUserActivatedHandler)

	mux.HandleFunc("POST /v1/tokens/authentication", a.postAuthenticationTokenHandler)

	return chain([]middleware{a.recoverPanic, a.rateLimit, a.authenticate, a.logRequest}, mux)
}
