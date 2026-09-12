package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tanvmn/mdb/internal/data"
	"github.com/tanvmn/mdb/internal/mailer"

	_ "github.com/lib/pq"
)

// version will later be generated at build time.
const version = "1.0.0"

func main() {
	a, err := newApp()
	a.handle(err)

	db, err := openDB(&a.conf)
	a.handle(err)
	defer db.Close()
	a.repo = data.NewRepo(db)

	a.handle(a.serve())
}

// handle slogs then os.Exit(1) if err != nil.
func (a *app) handle(err error) {
	if err != nil {
		a.log.Error(err.Error())
		os.Exit(1)
	}
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
	rate struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// app has dependencies for this service.
type app struct {
	conf
	log  *slog.Logger
	repo *data.Repo
	mail *mailer.Mailer
	wg   sync.WaitGroup
}

func newApp() (*app, error) {
	a := &app{
		log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}

	flag.IntVar(&a.port, "port", 4000, "Service's HTTP port")
	flag.StringVar(&a.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&a.db.dsn, "db-dsn", "", "PostgreSQL data source name")
	flag.IntVar(&a.db.maxOpenConns, "db-max-open-conns", 25, "Max open connections to PostgreSQL")
	flag.IntVar(&a.db.maxIdleConns, "db-max-idle-conns", 25, "Max idle connections to PostgreSQL")
	flag.DurationVar(&a.db.connMaxIdleTime, "db-conn-max-idle-time", 15*time.Minute, "Max idle time of connections to PostgreSQL")

	flag.StringVar(&a.smtp.host, "smtp-host", "smtp.gmail.com", "SMTP server's host")
	flag.IntVar(&a.smtp.port, "smtp-port", 587, "SMTP server's port")
	flag.StringVar(&a.smtp.username, "smtp-username", "", "SMTP client's login username")
	flag.StringVar(&a.smtp.password, "smtp-password", "", "SMTP client's login password")
	flag.StringVar(&a.smtp.sender, "smtp-sender", "mdb <no-reply@not.available.yet>", "SMTP sender")

	flag.Float64Var(&a.rate.rps, "rate-rps", 2, "Rate limit's max requests per second")
	flag.IntVar(&a.rate.burst, "rate-burst", 4, "Rate limit's max burst of requests")
	flag.BoolVar(&a.rate.enabled, "rate-enabled", true, "Enable rate limit")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(s string) error {
		a.cors.trustedOrigins = strings.Fields(s)

		return nil
	})

	flag.Parse()

	m, err := mailer.New(a.smtp.host, a.smtp.port, a.smtp.username, a.smtp.password, a.smtp.sender)
	if err != nil {
		return nil, err
	}
	a.mail = m

	return a, nil
}

// openDB establishes a connection pool to a database and pings it.
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
