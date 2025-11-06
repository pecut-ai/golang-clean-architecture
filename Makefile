.PHONY: help install build migrate-up migrate-down docker-up docker-down docker-restart run-api run-worker run clean

help:
	@echo "Available commands:"
	@echo "  make install        - Install Go dependencies"
	@echo "  make build          - Build API and Worker binaries"
	@echo "  make docker-up      - Start Docker containers"
	@echo "  make docker-down    - Stop Docker containers"
	@echo "  make docker-restart - Restart Docker containers"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make run-web        - Run WEB server"
	@echo "  make run-worker     - Run Worker"
	@echo "  make run            - Start containers, migrate, and run WEB + Worker"
	@echo "  make clean          - Clean build artifacts and stop containers"

install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build:
	@echo "Building binaries..."
	go build -o bin/web cmd/web/main.go
	go build -o bin/worker cmd/worker/main.go

docker-up:
	@echo "Starting Docker containers..."
	docker compose -f docker-compose.dev.yml up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker compose -f docker-compose.dev.yml down

docker-restart:
	@echo "Restarting Docker containers..."
	docker compose -f docker-compose.dev.yml restart

migrate-up:
	@echo "Running database migrations..."
	docker run --rm -v $(PWD)/db/migrations:/migrations --network host migrate/migrate \
		-path=/migrations/ \
		-database "postgresql://postgres:123@postgres:5432/golang_clean_architecture?sslmode=disable" up

migrate-down:
	@echo "Rolling back database migrations..."
	docker run --rm -v $(PWD)/db/migrations:/migrations --network host migrate/migrate \
		-path=/migrations/ \
		-database "postgresql://postgres:123@postgres:5432/golang_clean_architecture?sslmode=disable" down

run-web:
	@echo "Running WEB server..."
	go run cmd/web/main.go

run-worker:
	@echo "Running Worker..."
	go run cmd/worker/main.go

run:
	@echo "Starting containers and running application..."
	@make docker-up
# 	@echo "Waiting for containers to be ready..."
# 	@sleep 5
# 	@make migrate-up
# 	@echo "Starting API and Worker..."
# 	@make -j2 run-web run-worker

	@echo "Starting Docker containers..."
	docker compose -f docker-compose.dev.yml up

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	docker compose -f docker-compose.dev.yml down -v
