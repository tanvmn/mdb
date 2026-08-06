include .envrc

.PHONY: run/api

run/api:
	@go run ./cmd/api/ -db-dsn=${MDB_DB_DSN}
