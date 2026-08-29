package main

import (
	"context"
	"net/http"

	"github.com/tanvmn/mdb/internal/data"
)

// contextKey represents the type of http.Request's Context keys.
type contextKey string

const userContextKey = contextKey("user")

// contextSetUser sets a user in r.Context() and returns a new request whose context contains a user.
func (a *app) contextSetUser(r *http.Request, u *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, u)

	return r.WithContext(ctx)
}

// contextGetUser returns either a existing *data.User or data.AnonymousUser,
// so a panic occurs if no user is found.
func (a *app) contextGetUser(r *http.Request) *data.User {
	u, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request's context")
	}

	return u
}
