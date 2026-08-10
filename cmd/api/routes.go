package main

import "net/http"

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.notFound)

	mux.HandleFunc("GET /v1/health", a.getHealthHandler)

	mux.HandleFunc("GET /v1/movies/{id}", a.getMovieHandler)
	mux.HandleFunc("POST /v1/movies", a.postMovieHandler)
	mux.HandleFunc("PUT /v1/movies/{id}", a.putMovieHandler)
	mux.HandleFunc("DELETE /v1/movies/{id}", a.deleteMovieHandler)

	return chain([]middleware{a.recoverPanic, a.logRequest}, mux)
}
