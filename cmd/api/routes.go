package main

import "net/http"

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/health", a.getHealthHandler)
	mux.HandleFunc("POST /v1/movies", a.postMovieHandler)
	mux.HandleFunc("GET /v1/movies/{id}", a.getMovieHandler)

	return mux
}
