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
}

type conf struct {
	port int
	env  string
	db   struct {
		dsn             string
		maxOpenConns    int
		maxIdleConns    int
		maxConnIdleTime time.Duration
	}
}

// app has dependencies for this service.
type app struct {
	conf
	log *slog.Logger
}

func newApp() *app {
	a := &app{
		log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}

	flag.IntVar(&a.port, "port", 4000, "Service HTTP port")
	flag.StringVar(&a.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&a.db.dsn, "db-dsn", "", "PostgreSQL data source name")
	flag.IntVar(&a.db.maxOpenConns, "db-max-open-conns", 25, "Max open connections to PostgreSQL")
	flag.IntVar(&a.db.maxIdleConns, "db-max-idle-conns", 25, "Max idle connections to PostgreSQL")
	flag.DurationVar(&a.db.maxConnIdleTime, "db-max-conn-idle-time", 15*time.Minute, "Max idle time of connections to PostgreSQL")

	flag.Parse()

	_, err := openDB(&a.conf)
	if err != nil {
		a.log.Error(err.Error())
		os.Exit(1)
	}

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
