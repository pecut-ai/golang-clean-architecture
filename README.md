# Golang Clean Architecture Template

## Description

This is golang clean architecture template.

## Architecture

![Clean Architecture](architecture.png)

1. External system perform request (HTTP, gRPC, Messaging, etc)
2. The Delivery creates various Model from request data
3. The Delivery calls Use Case, and execute it using Model data
4. The Use Case create Entity data for the business logic
5. The Use Case calls Repository, and execute it using Entity data
6. The Repository use Entity data to perform database operation
7. The Repository perform database operation to the database
8. The Use Case create various Model for Gateway or from Entity data
9. The Use Case calls Gateway, and execute it using Model data
10. The Gateway using Model data to construct request to external system
11. The Gateway perform request to external system (HTTP, gRPC, Messaging, etc)

## Tech Stack

- Golang : https://github.com/golang/go
- Postgres (Database)
- Apache Kafka : https://github.com/apache/kafka

## Framework & Library

- GoFiber (HTTP Framework) : https://github.com/gofiber/fiber
- GORM (ORM) : https://github.com/go-gorm/gorm
- Viper (Configuration) : https://github.com/spf13/viper
- Golang Migrate (Database Migration) : https://github.com/golang-migrate/migrate
- Go Playground Validator (Validation) : https://github.com/go-playground/validator
- Logrus (Logger) : https://github.com/sirupsen/logrus
- Sarama (Kafka Client) : https://github.com/IBM/sarama

## Configuration

All configuration is in `.env` file. Copy and modify as needed:

```bash
# Application
APP_NAME=golang-clean-architecture
WEB_PORT=3000
WEB_PREFORK=false
LOG_LEVEL=4

# Database
DB_USERNAME=postgres
DB_PASSWORD=123
DB_HOST=localhost
DB_POST=5432
DB_NAME=golang_clean_architecture
DB_POOL_IDLE=10
DB_POOL_MAX=100
DB_POOL_LIFETIME=300

# Kafka
KAFKA_PRODUCER_ENABLED=false
KAFKA_BOOTSTRAP_SERVER=localhost:9092
KAFKA_GROUP_ID=golang-clean-arch-group
KAFKA_AUTO_OFFSET_RESET=earliest
```

## API Spec

All API Spec is in `api` folder.

## Quick Start

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make

### Setup and Run

```bash
# Install dependencies
make install

# Start everything (containers + migrations + app)
make run
```

This will:

1. Start Docker containers (Postgres & Kafka)
2. Run database migrations
3. Start both web server and worker

### Available Make Commands

```bash
make help              # Show all available commands
make install           # Install Go dependencies
make build             # Build binaries
make docker-up         # Start Docker containers
make docker-down       # Stop Docker containers
make docker-restart    # Restart Docker containers
make migrate-up        # Run database migrations
make migrate-down      # Rollback database migrations
make run-web           # Run web server only
make run-worker        # Run worker only
make run               # Start containers, migrate, and run everything
make clean             # Clean build artifacts and stop containers
```

## Database Migration

All database migration files are in `db/migrations` folder.

### Create Migration

```bash
migrate create -ext sql -dir db/migrations create_table_xxx
```

### Manual Migration

```bash
# Up
make migrate-up

# Down
make migrate-down
```

## Run Application

### Run unit test

```bash
go test -v ./test/
```

### Run web server only

```bash
make run-web
```

### Run worker only

```bash
make run-worker
```

### Run everything

```bash
make run
```
