package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

type middleware func(http.Handler) http.Handler

// chain chains each of mids by the order of their index, then chains them to final.
// chain panics if final or at least one of mids is nil.
func chain(mids []middleware, final http.Handler) http.Handler {
	isNil := func(m middleware) bool { return m == nil }
	if final == nil || slices.ContainsFunc(mids, isNil) {
		panic("chain: final or at least 1 middleware is nil")
	}
	for _, m := range slices.Backward(mids) {
		final = m(final)
	}

	return final
}

// recoverPanic defers recovering before calling next.ServeHTTP(w,r).
// If a panic occured, header "Connection: close" is set.
func (a *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// this makes http.Server close the conn after responding.
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

// rateLimit returns a handler that implements a rate limiting, which has a refill rate and a burst limit.
func (a *app) rateLimit(next http.Handler) http.Handler {
	if !a.limit.enable {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	type client struct {
		lastSeen time.Time
		lim      *rate.Limiter
	}
	var mu sync.Mutex
	clients := make(map[string]*client)

	// Delete any client that hasn't visit for more than 3min.
	// Remember, later, to use wrap this in a helper that handles panic.
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realip.FromRequest(r)
		mu.Lock()
		c, found := clients[ip]
		if !found {
			clients[ip] = &client{lim: rate.NewLimiter(rate.Limit(a.limit.rps), a.limit.burst)}
			c = clients[ip]
		}
		c.lastSeen = time.Now()

		if !c.lim.Allow() {
			mu.Unlock()
			a.tooManyRequests(w, r)
			return
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
