package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type wrap map[string]any

// id returns the positive integer value of parameter "id" in the path pattern registered in http.ServeMux.
func id(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id path parameter")
	}

	return id, nil
}

// json encodes data to JSON, set Content-Type: application/json, writes status, then writes the encoded data to w.
// If encoding fails, json calls a.serverErr.
// If writing to w fails, json serves status 500 without a body.
func (a *app) json(w http.ResponseWriter, r *http.Request, status int, data any) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		a.serverErr(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(bytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ok calls a.json with status 200.
func (a *app) ok(w http.ResponseWriter, r *http.Request, data any) {
	a.json(w, r, http.StatusOK, data)
}
