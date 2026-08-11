-include .env
export

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X golang-clean-architecture/internal/buildinfo.Version=$(VERSION) -X golang-clean-architecture/internal/buildinfo.Commit=$(COMMIT) -X golang-clean-architecture/internal/buildinfo.BuildTime=$(BUILD_TIME)
DATABASE_URL ?= postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: help install build openapi test migrate-up migrate-down run-web run-worker run clean

help:
	@echo "Available commands:"
	@echo "  make install        - Install Go dependencies"
	@echo "  make build          - Build API and Worker binaries"
	@echo "  make openapi        - Regenerate api/openapi.json"
	@echo "  make test           - Run focused package tests"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make run-web        - Run WEB server"
	@echo "  make run-worker     - Run Worker"
	@echo "  make run            - Run the web server"
	@echo "  make clean          - Clean build artifacts"

install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build:
	@echo "Building binaries..."
	go build -ldflags "$(LDFLAGS)" -o bin/web ./cmd/web
	go build -ldflags "$(LDFLAGS)" -o bin/worker ./cmd/worker

openapi:
	go run ./cmd/open-api api/openapi.json

test:
	go test -p=1 ./internal/config ./internal/logging ./internal/middleware ./internal/delivery/http/route ./internal/bootstrap

migrate-up:
	@echo "Running database migrations..."
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path db/migrations -database "$(DATABASE_URL)" down

run-web:
	@echo "Running WEB server..."
	go run cmd/web/main.go

run-worker:
	@echo "Running Worker..."
	go run cmd/worker/main.go

run:
	@$(MAKE) run-web

clean:
	@echo "Cleaning up..."
	rm -rf ./bin
