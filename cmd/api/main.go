package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/tanvmn/mdb/internal/data"
)

// version will later be generated at build time.
const version = "1.0.0"

func main() {
	a := newApp()

	db, err := openDB(&a.conf)
	a.handle(err)
	defer db.Close()

	a.repo = data.NewRepo(db)

	a.handle(a.serve())
}

// conf has configurable settings of app.
type conf struct {
	port int
	env  string
	db   struct {
		dsn             string
		maxOpenConns    int
		maxIdleConns    int
		connMaxIdleTime time.Duration
	}
	limit struct {
		rps    float64
		burst  int
		enable bool
	}
}

// app has dependencies for this service.
type app struct {
	conf
	log  *slog.Logger
	repo *data.Repo
}

func newApp() *app {
	a := &app{
		log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}

	flag.IntVar(&a.port, "port", 9000, "Service's HTTP port")
	flag.StringVar(&a.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&a.db.dsn, "db-dsn", "", "PostgreSQL data source name")
	flag.IntVar(&a.db.maxOpenConns, "db-max-open-conns", 25, "Max open connections to PostgreSQL")
	flag.IntVar(&a.db.maxIdleConns, "db-max-idle-conns", 25, "Max idle connections to PostgreSQL")
	flag.DurationVar(&a.db.connMaxIdleTime, "db-conn-max-idle-time", 15*time.Minute, "Max idle time of connections to PostgreSQL")

	flag.Float64Var(&a.limit.rps, "ratelimit-rps", 2, "Rate limit's max requests per second")
	flag.IntVar(&a.limit.burst, "ratelimit-burst", 4, "Rate limit's max burst of requests")
	flag.BoolVar(&a.limit.enable, "ratelimit-enable", true, "Enable rate limit")

	flag.Parse()

	return a
}

// handle slogs then exit(1) if err != nil
func (a *app) handle(err error) {
	if err != nil {
		a.log.Error(err.Error())
		os.Exit(1)
	}
}

// openDB establishes a connection pool and pings db.
func openDB(c *conf) (*sql.DB, error) {
	db, err := sql.Open("postgres", c.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(c.db.maxOpenConns)
	db.SetMaxIdleConns(c.db.maxIdleConns)
	db.SetConnMaxIdleTime(c.db.connMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
