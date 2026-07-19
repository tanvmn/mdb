package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func (a *app) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf("localhost:%v", a.port),
		Handler:      a.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(a.log.Handler(), slog.LevelError),
	}
	a.log.Info("listenning", "addr", srv.Addr, "env", a.env)

	return srv.ListenAndServe()
}
