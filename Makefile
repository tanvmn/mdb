include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

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


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## build/api: build cmd/api application.
.PHONY: build/api
build/api:
	@go build -o=./bin/api -ldflags='-s' ./cmd/api
	@GOOS=linux GOARCH=amd64 go build -o=./bin/linux_amd64/api -ldflags='-s' ./cmd/api

## run/api: run cmd/api application.
.PHONY: run/api
run/api: build/api
	@./bin/api -version=$${v:-false} -db-dsn=${DB_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

## run/linux_amd64/api: run cmd/api application binary for linux/amd64.
.PHONY: run/linux_amd64/api
run/linux_amd64/api: build/api
	@./bin/linux_amd64/api -version=$${v:-false} -db-dsn=${DB_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD} -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

## db/migrate/create name=$1: create a new pair of sequential db migration files.
.PHONY: db/migrate/create
db/migrate/create:
	@migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrate/force version=$1: force db's state to a version, usually to fix migration error.
.PHONY: db/migrate/force
db/migrate/force:
	@migrate -path=./migrations/ -database=${DB_DSN} force ${version} # syntax: migrate/force version=<version number>

## db/migrate/goto version=$1: migrate up or down to a version.
.PHONY: db/migrate/goto
db/migrate/goto:
	@migrate -path=./migrations/ -database=${DB_DSN} goto ${version}

## db/migrate/up: migrate up to the highest version.
.PHONY: db/migrate/up
db/migrate/up: confirm
	@migrate -path=./migrations/ -database=${DB_DSN} up

## db/migrate/down: migrate down to the state before migrating.
.PHONY: db/migrate/down
db/migrate/down:
	@migrate -path=./migrations/ -database=${DB_DSN} down

## db/migrate/version: check migrate version.
.PHONY: db/migrate/version
db/migrate/version:
	@migrate -path=./migrations/ -database=${DB_DSN} version

## db/start: start db.
.PHONY: db/start
db/start:
	@sudo systemctl start postgresql.service
## db/stop: stop db.
.PHONY: db/stop
db/stop:
	@sudo systemctl stop postgresql.service
	@sudo systemctl status postgresql.service

## psql: start psql.
.PHONY: psql
psql: db/start
	@psql ${DB_DSN}


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: tidy module dependencies and format all .go files.
.PHONY: tidy
tidy:
	go mod tidy
	go mod verify
	go mod vendor
	go fmt ./...

## audit: quality control.
.PHONY: audit
audit:
	go mod tidy -diff
	go mod verify
	go vet ./...
	go tool staticcheck ./...

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

prod_host_ip = '168.144.249.126'

## prod/connect: connect to the production server
.PHONY: prod/connect
prod/connect:
	ssh mdb@${prod_host_ip}
