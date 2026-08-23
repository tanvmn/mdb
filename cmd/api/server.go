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
	// - create a new goroutine to wait for one of signals: syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT
	// - when signal is received, call srv.Shutdown(30s timeout context) and send the potential error to a chan error (let's call shutdownErr)
	// - with the request's goroutine, after srv.Shutdown's called in other goroutine:
	//	-- if srv.ListenAndServe() doesn't return http.ErrServerClosed, then srv stops listenning UNSUCCESSFULLY
	//		--- return err
	//	if srv.ListenAndServe() returns http.ErrServerClosed from srv.Shutdown(), then srv stops listenning SUCCESSFULLY
	// 		--- receive the potential error from srv.Shutdown() via a chan shutdownErr. If err != nil, return err, else server's graceful shutdown is successful.

	shutdownErr := make(chan error)

	a.godo(func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) // seems to block until receiving sig.

		a.log.Info("shutting down server", "signal", (<-quit).String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)
	})

	a.log.Info("listenning", "addr", srv.Addr, "env", a.env)

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	if err := <-shutdownErr; err != nil {
		return err
	}

	a.log.Info("waiting for other goroutines", "addr", srv.Addr)
	a.wg.Wait()

	a.log.Info("server stopped", "addr", srv.Addr)

	return nil
}
