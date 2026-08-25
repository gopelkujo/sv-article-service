# Article Service — common developer tasks.
# Copy .env.example to .env before running migrate / run targets.

.PHONY: run build test test-integration lint migrate-up migrate-down migrate-create docker-up docker-down docker-up-all tidy fmt tools

ifneq (,$(wildcard .env))
    include .env
    export
endif

APP_NAME ?= article-service
MIGRATE ?= migrate
MIGRATIONS_DIR ?= ./migrations
DATABASE_URL ?= mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?multiStatements=true

## run: start the API server (loads .env via godotenv)
run:
	go run ./cmd/api

## build: compile the API binary into ./bin/api
build:
	mkdir -p bin
	go build -o bin/api ./cmd/api

## test: run all unit and integration tests
test:
	go test ./... -count=1 -race -timeout 120s

## test-integration: run tests that require a live MySQL (uses .env)
test-integration:
	RUN_DB_TESTS=1 go test ./... -count=1 -race -timeout 120s

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## tidy: sync go.mod / go.sum
tidy:
	go mod tidy

## fmt: format all Go sources
fmt:
	gofmt -w .
	goimports -w .

## migrate-up: apply all pending migrations to the article database
migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

## migrate-down: roll back the most recent migration
migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

## migrate-create: create a new migration pair (usage: make migrate-create name=add_foo)
migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<migration_name>" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## docker-up: start MySQL (and optionally the API) via docker-compose
docker-up:
	docker compose up -d mysql

## docker-down: stop and remove compose services
docker-down:
	docker compose down

## docker-up-all: start MySQL + API containers
docker-up-all:
	docker compose up -d --build

## tools: install local CLI dependencies (migrate + golangci-lint)
tools:
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.64.8
