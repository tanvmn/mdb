include .envrc

## help: print this help message.
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## help/run: print flags for running cmd/api application.
.PHONY: help/run
help/run:
	@go run ./cmd/api/ -help

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/n] ' && read ans && [ $${ans:-n} = y ]

## run/api: run cmd/api application.
.PHONY: run/api
run/api:
	@go run ./cmd/api/ -db-dsn=${DB_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

.PHONY: migrate/version, migrate/up, migrate/down, migrate/goto
## db/migrate/create name=$1: create a new pair of sequential db migration files.
db/migrate/create:
	@migrate create -seq -ext=.sql -dir=./migrations ${name}
## db/migrate/force version=$1: force db's state to a version, usually to fix migration error.
db/migrate/force:
	@migrate -path=./migrations/ -database=${DB_DSN} force ${version} # syntax: migrate/force version=<version number>
## db/migrate/goto version=$1: migrate up or down to a version.
db/migrate/goto:
	@migrate -path=./migrations/ -database=${DB_DSN} goto ${version}
## db/migrate/up: migrate up to the highest version.
db/migrate/up: confirm
	@migrate -path=./migrations/ -database=${DB_DSN} up
## db/migrate/down: migrate down to the state before migrating.
db/migrate/down:
	@migrate -path=./migrations/ -database=${DB_DSN} down
## db/migrate/version: check migrate version.
db/migrate/version:
	@migrate -path=./migrations/ -database=${DB_DSN} version

.PHONY: db/start, db/stop
## db/start: start db.
db/start:
	@sudo systemctl start postgresql.service
## db/stop: stop db.
db/stop:
	@sudo systemctl stop postgresql.service
	@sudo systemctl status postgresql.service

.PHONY: psql
## psql: start psql.
psql: db/start
	@psql ${DB_DSN}
