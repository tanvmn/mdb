package main

import "net/http"

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.notFound)

	mux.HandleFunc("GET /v1/health", a.getHealthHandler)

	mux.HandleFunc("GET /v1/movies", a.validateActivated(a.getMoviesHandler))
	mux.HandleFunc("POST /v1/movies", a.validateActivated(a.postMovieHandler))
	mux.HandleFunc("GET /v1/movies/{id}", a.validateActivated(a.getMovieHandler))
	mux.HandleFunc("PATCH /v1/movies/{id}", a.validateActivated(a.patchMovieHandler))
	mux.HandleFunc("DELETE /v1/movies/{id}", a.validateActivated(a.deleteMovieHandler))

	mux.HandleFunc("POST /v1/users", a.postUserHandler)
	mux.HandleFunc("PUT /v1/users/activated", a.putUserActivatedHandler)

	mux.HandleFunc("POST /v1/tokens/authentication", a.postAuthenticationTokenHandler)

	return chain([]middleware{a.recoverPanic, a.rateLimit, a.authenticate, a.logRequest}, mux)
}
