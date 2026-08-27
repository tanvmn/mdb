include .envrc

.PHONY: run/api
run/api:
	@go run ./cmd/api/ -db-dsn=${DB_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD}

.PHONY: help
help:
	@go run ./cmd/api/ -help

.PHONY: migrate/version, migrate/up, migrate/down, migrate/goto
migrate/version:
	@migrate -path=./migrations/ -database=${DB_DSN} version
migrate/up:
	@migrate -path=./migrations/ -database=${DB_DSN} up
migrate/down:
	@migrate -path=./migrations/ -database=${DB_DSN} down
migrate/goto:
	@migrate -path=./migrations/ -database=${DB_DSN} goto ${V}
