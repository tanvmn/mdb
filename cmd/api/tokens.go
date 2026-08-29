package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/validator"
)

func (a *app) postAuthenTokenHandler(w http.ResponseWriter, r *http.Request) {
	var i struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := decodeJSON(w, r.Body, &i); err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	v.Validate(i.Email != nil, "email", "must be provided")
	v.Validate(i.Password != nil, "password", "must be provided")
	if !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}
	data.ValidateEmail(v, *i.Email)
	data.ValidatePassword(v, *i.Password)
	if !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	u, err := a.repo.Users.GetByEmail(*i.Email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.invalidCredentials(w, r)
		} else {
			a.serverErr(w, r, err)
		}
		return
	}
	if err = u.Password.Matches(*i.Password); err != nil {
		if errors.Is(err, data.ErrMismatchedPassword) {
			a.invalidCredentials(w, r)
		} else {
			a.serverErr(w, r, err)
		}
		return
	}

	if err = a.repo.Tokens.DeleteAllForUser(data.ScopeAuthentication, u.ID); err != nil {
		a.serverErr(w, r, err)
		return
	}
	t, err := a.repo.Tokens.New(u.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	a.created(w, r, wrap{"authenticationToken": t})
}
