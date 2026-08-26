package main

import (
	"context"
	"net/http"

	"github.com/tanvmn/mdb/internal/data"
)

// contextKey represents the type of http.Request's Context keys.
type contextKey string

const userContextKey = contextKey("user")

func (a *app) contextSetUser(r *http.Request, u *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, u)

	return r.WithContext(ctx)
}

// contextGetUser panics if user isn't found.
func (a *app) contextGetUser(r *http.Request) *data.User {
	u, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request's context")
	}

	return u
}
