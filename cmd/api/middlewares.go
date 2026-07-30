package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
)

type middleware func(http.Handler) http.Handler

// chain recursively calls http.Handler.ServeHTTP(w,r) returned by each of mids by their index order, then calls final.
// chain panics if final == nil or at least one of mids is nil.
func chain(mids []middleware, final http.Handler) http.Handler {
	isNil := func(m middleware) bool { return m == nil }
	if final == nil || slices.ContainsFunc(mids, isNil) {
		panic("chain: final or at least 1 middleware == nil")
	}
	for _, m := range slices.Backward(mids) {
		final = m(final)
	}

	return final
}

// recoverPanic defer a recover before calling next.ServeHTTP(w,r).
// If a panic occured, header "Connection: close" is set.
func (a *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// this triggers http.Server to close the conn after responsing.
				w.Header().Set("Connection", "close")
				a.serverErr(w, r, fmt.Errorf("%v", err))
				debug.PrintStack()
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (a *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.log.Info("request", "addr", r.RemoteAddr, "method", r.Method, "uri", r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}
