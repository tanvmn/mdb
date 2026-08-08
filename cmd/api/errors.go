package main

import (
	"fmt"
	"net/http"
)

type respErr struct {
	Status  string `json:"status"`
	Message any    `json:"message"`
}

func (a *app) logErr(r *http.Request, err error) {
	a.log.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
}

// err calls a.json with a respErr value as data.
func (a *app) err(w http.ResponseWriter, r *http.Request, status int, message any) {
	a.json(w, r, status, respErr{
		Status:  fmt.Sprintf("%v %v", status, http.StatusText(status)),
		Message: message,
	})
}

// serverErr calls a.err with status 500 and "server error" as message.
func (a *app) serverErr(w http.ResponseWriter, r *http.Request, err error) {
	a.logErr(r, err)
	a.err(w, r, http.StatusInternalServerError, "server error")
}

// notFound calls a.err with status 404 and "resource not found" as message.
func (a *app) notFound(w http.ResponseWriter, r *http.Request) {
	a.err(w, r, http.StatusNotFound, "resource not found")
}

// badRequest calls a.err with status 400 and err.Error() as message.
func (a *app) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	a.err(w, r, http.StatusBadRequest, err.Error())
}

// unprocessableEntity calls a.json with status 422.
func (a *app) unprocessableEntity(w http.ResponseWriter, r *http.Request, message any) {
	a.err(w, r, http.StatusUnprocessableEntity, message)
}
