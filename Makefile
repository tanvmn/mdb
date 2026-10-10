include .envrc

.PHONY: run/api
run/api:
	@go run ./cmd/api/ -db-dsn=${DB_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

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
	@migrate -path=./migrations/ -database=${DB_DSN} goto ${V} # export V before use
migrate/force:
	@migrate -path=./migrations/ -database=${DB_DSN} force ${V}

.PHONY: db/start, db/stop
db/start:
	@sudo systemctl start postgresql.service
db/stop:
	@sudo systemctl stop postgresql.service
	@sudo systemctl status postgresql.service

.PHONY: psql
psql: db/start
	@psql ${DB_DSN}
