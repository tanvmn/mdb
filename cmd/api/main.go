package main

import (
	"flag"
	"log/slog"
	"os"
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
	flag.Parse()

	return a
}
