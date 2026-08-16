package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// graceful shutdown flow:
	// - create a new goroutine to wait for one of syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT
	// - when signal is received, call srv.Shutdown(30s timeout context) and send the potential error to a chan error
	// - with the main goroutine, after calling srv.Shutdown:
	//	if srv.ListenAndServe() doesn't return http.ErrServerClosed, then srv stops listenning UNSUCCESSFULLY; so return err
	//	if srv.ListenAndServe() returns http.ErrServerClosed from srv.Shutdown, then srv stops listenning successfully, and move to the step below
	// - if this step is reached, then receive the potential error from srv.Shutdown. If err != nil, return err, else server's graceful shutdown is successful.

	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) // seems to block until receiving sig.

		a.log.Info("shutting down server", "signal", (<-quit).String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	}()

	a.log.Info("listenning", "addr", srv.Addr, "env", a.env)

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	if err := <-shutdownErr; err != nil {
		return err
	}

	a.log.Info("server shutdown successful", "addr", srv.Addr)

	return nil
}
