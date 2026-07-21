package main

import (
	"fmt"
	"net/http"
)

func (a *app) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := id(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "movie with ID %v", id)
}

func (a *app) postMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "post movie")
}
