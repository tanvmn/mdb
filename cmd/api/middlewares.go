package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"strconv"
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
				w.Header().Add("Connection", "close")
				a.serverErr(w, r, fmt.Errorf("%v", err))
				debug.PrintStack()
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (a *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.log.Info("request", "addr", realip.FromRequest(r), "method", r.Method, "uri", r.URL.RequestURI())

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

	// Every minute, delete any client that hasn't visited for more than 3min.
	go a.do(func() {
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
		// w.Header().Add("X-RateLimit-Remaining", fmt.Sprintf("%v", math.Floor(c.lim.Tokens())))
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// authenticate sets "Vary: Authorization", then authenticates based on "Authorization: Bearer <plaintext token>".
func (a *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Indicate that a response may vary based on a request's Authorization header, thus the response shouldn't be cached.
		w.Header().Add("Vary", "Authorization")

		// Create and set an anonymous user is better than not setting an user when auth token isn't found,
		// cause then authenticate can be chained before mux,
		// whereas with not setting, authenticate needs to be prefixed with each needed handler.
		// Plus, no need to check if *User exists everytime getting it from req's context
		// cause there'll always be either a real or an anonymous user.
		auth := r.Header.Get("Authorization")
		if auth == "" {
			r = a.setContextUser(r, data.AnonymousUser)
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
		r = a.setContextUser(r, u)

		next.ServeHTTP(w, r)
	})
}

// validateUserAuthenticated only calls next if the request's user is authenticated.
func (a *app) validateUserAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.contextUser(r).IsAnonymous() {
			a.unauthorized(w, r, "user needs to be authenticated")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateUserActivated only calls next if the request's user is authenticated and activated.
func (a *app) validateUserActivated(next http.HandlerFunc) http.HandlerFunc {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.contextUser(r).Activated {
			a.forbidden(w, r, "user needs to be activated")
			return
		}

		next.ServeHTTP(w, r)
	})

	return a.validateUserAuthenticated(f)
}

// authorize only calls next if the authenticated and activated user in request's context has perm.
func (a *app) authorize(perm string, next http.HandlerFunc) http.HandlerFunc {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := a.repo.Users.HasPerm(a.contextUser(r).ID, perm); err != nil {
			if errors.Is(err, data.ErrNotHavePerm) {
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

func (a *app) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rule of thumb: if response's content is based on value of a request header,
		// then include that request header (the name not the value) in the response's Vary header, even if the request didn't send that header.
		w.Header().Add("Vary", "Origin")
		// Response'll be different if the request includes this header, cause then the request could be preflight.
		w.Header().Add("Vary", "Access-Control-Request-Method")

		if origin := r.Header.Get("Origin"); slices.Contains(a.cors.trustedOrigins, origin) {
			w.Header().Add("Access-Control-Allow-Origin", origin)
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				// Try not to set these headers to * (wildcard), cause now only 77% of browsers support this, any don't will block.
				// Authorization must be explicitly added like: "Access-Control-Allow-Headers: Authorization, *".
				// For preflights with credentials (cookie, basic auth), browsers will interpret "*" in the responses as a literal string instead of a wildcard,
				// so using wildcard won't work.
				w.Header().Add("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE") // no need to include CORS-safe methods
				w.Header().Add("Access-Control-Allow-Headers", "Authorization, Content-Type") // no need to include CORS-safe or forbidden headers
				w.WriteHeader(http.StatusOK)                                                  // prevent automatically setting 204 No Content
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

type metricsResponseWriter struct {
	w             http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		w:          w,
		statusCode: http.StatusOK, // in case no handlers call w.WriteHeader(), w.Write()
	}
}

func (mw *metricsResponseWriter) Header() http.Header {
	return mw.w.Header()
}

func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	// this call isn't inside if to let w.WriteHeader() be.
	mw.w.WriteHeader(statusCode)
	if !mw.headerWritten {
		mw.statusCode = statusCode
		mw.headerWritten = true
	}
}

func (mw *metricsResponseWriter) Write(data []byte) (int, error) {
	n, err := mw.w.Write(data)
	if err == nil {
		mw.headerWritten = true // w.Write will either set 200 or preserve the result of w.WriteHeader()
	}

	return n, err
}

func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.w
}

func metrics(next http.Handler) http.Handler {
	requestsReceived := expvar.NewInt("requestsReceived")
	responsesSent := expvar.NewInt("responsesSent")
	responsesSentByStatus := expvar.NewMap("responsesSentByStatus")
	processingTimeMicrosec := expvar.NewInt("processingTimeMicrosec")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestsReceived.Add(1)

		mw := newMetricsResponseWriter(w)
		next.ServeHTTP(mw, r)

		responsesSent.Add(1)
		responsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)
		processingTimeMicrosec.Add(time.Since(start).Microseconds())
	})
}
