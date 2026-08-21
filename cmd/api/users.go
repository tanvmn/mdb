package main

import (
	"errors"
	"net/http"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/validator"
)

func (a *app) postUserHandler(w http.ResponseWriter, r *http.Request) {
	var i struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := decodeJSON(w, r.Body, &i); err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	v.Validate(i.Name != nil, "name", "must be provided")
	v.Validate(i.Email != nil, "email", "must be provided")
	v.Validate(i.Password != nil, "password", "must be provided")
	if data.ValidatePassword(v, *i.Password); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	u := &data.User{
		Name:      *i.Name,
		Email:     *i.Email,
		Activated: false,
	}
	if err := u.Password.Set(*i.Password); err != nil {
		a.serverErr(w, r, err)
		return
	}
	if data.ValidateUser(v, u); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	if err := a.repo.Users.Add(u); err != nil {
		if errors.Is(err, data.ErrDuplicateEmail) {
			v.AddErr("email", "already used, try another")
			a.unprocessableEntity(w, r, v.Errors)
		} else {
			a.serverErr(w, r, err)
		}
		return
	}

	a.wgo(func() {
		if err := a.mail.Send(u.Email, "user_welcome.tmpl.html", u); err != nil {
			a.logErr(r, err)
		}
	})

	// this means the request has been accepted and still being processed.
	a.accepted(w, r, wrap{"user": u})
}
