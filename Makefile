include .envrc

.PHONY: run/api
run/api:
	@go run ./cmd/api/ -db-dsn=${DB_DSN}

.PHONY: help
help:
	@go run ./cmd/api/ -help
