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

// conflict calls a.err with status 409.
func (a *app) conflict(w http.ResponseWriter, r *http.Request, message any) {
	a.err(w, r, http.StatusConflict, message)
}

// editConflict calls a.conflict with "edit conflict occurred, try again" as message.
func (a *app) editConflict(w http.ResponseWriter, r *http.Request) {
	a.conflict(w, r, "edit conflict occurred, try again")
}

// tooManyRequests calls a.err with status 429, and "exceeded rate limit, retry later" as message.
func (a *app) tooManyRequests(w http.ResponseWriter, r *http.Request) {
	a.err(w, r, http.StatusTooManyRequests, "exceeded rate limit, retry later")
}

// unauthorized calls a.err with status 401.
func (a *app) unauthorized(w http.ResponseWriter, r *http.Request, message any) {
	a.err(w, r, http.StatusUnauthorized, message)
}

// forbidden calls a.err with status 403.
func (a *app) forbidden(w http.ResponseWriter, r *http.Request, message any) {
	a.err(w, r, http.StatusForbidden, message)
}

// invalidCredentials calls a.unauthorized with "invalid authentication credentials" as message.
func (a *app) invalidCredentials(w http.ResponseWriter, r *http.Request) {
	a.unauthorized(w, r, "invalid authentication credentials")
}

// invalidAuthenToken set "WWW-Authenticate: Bearer", then calls a.unauthorized with "invalid or missing authentication token" as message.
func (a *app) invalidAuthenToken(w http.ResponseWriter, r *http.Request) {
	// inform clients that server expects authentications using Bearer token.
	w.Header().Set("WWW-Authenticate", "Bearer")
	a.unauthorized(w, r, "invalid or missing authentication token")
}
