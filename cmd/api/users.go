package main

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/validator"
)

const tokenLifetime = 3 * 24 * time.Hour

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

	t, err := a.repo.Tokens.New(u.ID, tokenLifetime, data.ScopeActivation)
	if err != nil {
		a.serverErr(w, r, err)
		return
	}

	a.wgo(func() {
		data := map[string]any{
			"activationToken": t.Plaintext,
			"userID":          u.ID,
			"tokenLifetime":   math.Floor(tokenLifetime.Hours() / 24),
		}
		if err := a.mail.Send(u.Email, "user_welcome.tmpl.html", data); err != nil {
			a.logErr(r, err)
		}
	})

	// this means the request has been accepted and still being processed.
	a.accepted(w, r, wrap{"user": u})
}

// user calls a.ok with wrap{"user": u} as data.
func (a *app) user(w http.ResponseWriter, r *http.Request, u *data.User) {
	a.ok(w, r, wrap{"user": u})
}

func (a *app) putUserActivatedHandler(w http.ResponseWriter, r *http.Request) {
	var i struct {
		Token *string `json:"token"`
	}
	if err := decodeJSON(w, r.Body, &i); err != nil {
		a.badRequest(w, r, err)
		return
	}

	v := validator.New()
	v.Validate(i.Token != nil, "token", "must be provided")
	if data.ValidateToken(v, *i.Token); !v.Valid() {
		a.unprocessableEntity(w, r, v.Errors)
		return
	}

	u, err := a.repo.Users.GetForToken(data.ScopeActivation, *i.Token)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			v.AddErr("token", "invalid or expired activation token")
			a.unprocessableEntity(w, r, v.Errors)
		} else {
			a.serverErr(w, r, err)
		}
		return
	}
	u.Activated = true

	if err = a.repo.Users.Set(u); err != nil {
		if errors.Is(err, data.ErrEditConflict) {
			a.editConflict(w, r)
		} else {
			a.serverErr(w, r, err)
		}
		return
	}
	if err = a.repo.Tokens.DeleteAllForUser(data.ScopeActivation, u.ID); err != nil {
		a.serverErr(w, r, err)
		return
	}

	a.user(w, r, u)
}
