package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/validator"
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
		a.log.Info("request", "addr", r.RemoteAddr, "method", r.Method, "uri", r.URL.RequestURI(), "userID", a.contextGetUser(r).ID)

		next.ServeHTTP(w, r)
	})
}

// rateLimit returns a handler that implements a rate limiting, which has a refill rate and a burst limit.
func (a *app) rateLimit(next http.Handler) http.Handler {
	if !a.rate.enabled {
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

	// Every minute, delete any client that hasn't visit for more than 3min.
	a.godo(func() {
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
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realip.FromRequest(r)
		mu.Lock()
		c, found := clients[ip]
		if !found {
			clients[ip] = &client{lim: rate.NewLimiter(rate.Limit(a.rate.rps), a.rate.burst)}
			c = clients[ip]
		}
		c.lastSeen = time.Now()

		if !c.lim.Allow() {
			mu.Unlock()
			a.tooManyRequests(w, r)
			return
		}
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%v", math.Floor(c.lim.Tokens())))
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// authenticate sets "Vary: Authorization", then authenticates based on "Authorization: Bearer <token>".
func (a *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inform clients who cache, that a response may vary based on a request's Authorization header.
		w.Header().Set("Vary", "Authorization")

		// Create and set an anonymous user is better than not setting an user when authentication fails,
		// cause then authen middleware can be chained before mux,
		// whereas with not setting, authen needs to be prefixed with each needed handler.
		// Plus, no need to check if *User exists everytime getting it from req's context
		// cause there'll always be either a real or an anonymous user.
		auth := r.Header.Get("Authorization")
		if auth == "" {
			r = a.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			a.invalidAuthenToken(w, r)
			return
		}
		token := parts[1]
		v := validator.New()
		if data.ValidateToken(v, token); !v.Valid() {
			a.invalidAuthenToken(w, r)
			return
		}

		u, err := a.repo.Users.GetByToken(data.ScopeAuthentication, token)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				a.invalidAuthenToken(w, r)
			} else {
				a.serverErr(w, r, err)
			}
			return
		}
		r = a.contextSetUser(r, u)

		next.ServeHTTP(w, r)
	})
}

// validateUserAuthenticated only calls next if the request's user is authenticated.
func (a *app) validateUserAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := a.contextGetUser(r)
		if u.IsAnonymous() {
			a.unauthorized(w, r, "resource requires authentication")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateUserActivated only calls next if the request's user is authenticated and activated.
func (a *app) validateUserActivated(next http.HandlerFunc) http.HandlerFunc {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := a.contextGetUser(r)
		if !u.Activated {
			a.forbidden(w, r, "resource requires user to be activated")
			return
		}

		next.ServeHTTP(w, r)
	})

	return a.validateUserAuthenticated(f)
}

func (a *app) authorize(code string, next http.HandlerFunc) http.HandlerFunc {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ps, err := a.repo.Perms.GetAllForUser(a.contextGetUser(r).ID)
		// if err != nil {
		// 	a.serverErr(w, r, err)
		// 	return
		// }
		// if !ps.Include(code) {
		// 	a.forbidden(w, r, "user doesn't have permission to access this resource")
		// 	return
		// }

		if err := a.repo.Users.HasPerm(a.contextGetUser(r).ID, code); err != nil {
			if errors.Is(err, data.ErrUserNotHavePerm) {
				a.forbidden(w, r, "user doesn't have permission to access this resource")
			} else {
				a.serverErr(w, r, err)
			}
			return
		}

		next.ServeHTTP(w, r)
	})

	return a.validateUserActivated(f)
}
