package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// version will later be generated at build time.
const version = "1.0.0"

func main() {
	a := newApp()
	if err := a.serve(); err != nil {
		a.log.Error(err.Error())
		os.Exit(1)
	}
	defer a.data.Close()
}

type conf struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

// app has dependencies for this service.
type app struct {
	conf
	log  *slog.Logger
	data *sql.DB
}

func newApp() *app {
	a := &app{
		log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}

	flag.IntVar(&a.port, "port", 4000, "Service HTTP port")
	flag.StringVar(&a.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&a.db.dsn, "db-dsn", "", "PostgreSQL Data Source Name")

	flag.Parse()

	db, err := openDB(&a.conf)
	if err != nil {
		a.log.Error(err.Error())
		os.Exit(1)
	}
	a.data = db

	return a
}

func openDB(c *conf) (*sql.DB, error) {
	db, err := sql.Open("postgres", c.db.dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
