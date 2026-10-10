package main

import (
	"net/http"
)

func (a *app) getHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(201)
	w.WriteHeader(202)
	return
	health := map[string]any{
		"status": "available",
		"systemInfo": map[string]any{
			"environment": a.env,
			"version":     version,
		},
	}
	a.ok(w, r, health)
}
