package main

import (
	"net/http"
)

func (a *app) getHealthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]any{
		"status": "available",
		"systemInfo": map[string]any{
			"environment": a.env,
			"version":     version,
		},
	}
	// fmt.Println(r.Header.Get("H"))
	a.ok(w, r, health)
}
