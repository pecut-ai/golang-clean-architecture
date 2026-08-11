# Go Clean Architecture Service Template

This repository is a production-oriented starting point for a Fiber + Huma service. It keeps the example contact/address domain small while demonstrating the runtime seams that every service needs: validated environment configuration, structured request-correlated logging, auth-service integration, OpenAPI-backed handlers, and deterministic startup/shutdown ownership.

## What is included

- Viper-backed `.env` and process-environment loading into typed configuration.
- Startup validation with joined errors for invalid or missing settings.
- Zap structured logging with contextual `request_id`, request metadata, user identity, and component names.
- A GORM logger adapter using the same configured logger and request context.
- Signal-driven graceful shutdown and panic recovery at process and HTTP boundaries.
- Build metadata in `internal/buildinfo`; `APP_VERSION` is intentionally not used.
- Auth-service v2 middleware and downstream request-ID propagation over gRPC.
- CORS, trusted-proxy, request-ID, request logging, recovery, and auth middleware in an explicit order.
- Real Huma handlers: runtime routing, validation, responses, and OpenAPI are defined once.

The current PostgreSQL handle is intentionally a transitional single-database seam. The next phase can replace `config.OpenDatabase` and the `bootstrap.Application.DB` field with the custom multi-DB manager without changing handler ownership or falling back to non-GORM queries.

## Runtime flow

```text
cmd/web
  -> config.Load (Viper + validation)
  -> logging.New
  -> bootstrap.NewApplication
       -> Fiber and global middleware
       -> GORM handle with component=gorm logger
       -> optional Kafka producer
       -> auth-service v2 client and middleware
       -> Huma API and real example handlers
  -> listen
  -> signal or server failure
  -> bounded graceful shutdown
```

Middleware order matters. Request ID and request context run first, the access logger wraps the whole request, panic recovery wraps downstream work, CORS runs before route handling, and auth protects every `/api` route except `/api/health`. Authenticated user data is then added to the shared logging context before handlers run.

Every HTTP response includes `X-Request-ID`. Every completed request emits one `http_request` event containing at least `request_id`, `component`, method, path, status, latency, and response size. Internal handlers, auth-service calls, and GORM queries reuse that context.

## Configuration

Copy the annotated example and replace secrets:

```bash
cp .env.example .env
```

[`.env.example`](.env.example) documents every supported setting and its expected format. `APP_NAME`, `APP_ENV`, `APP_URL`, and `APP_PORT` are required and come from the environment. Process environment values override the dotenv file. Use `ENV_FILE=/path/to/file.env` to select another file.

Important validation rules:

- `APP_ENV` is one of `development`, `test`, `staging`, or `production`.
- `APP_URL` is an absolute URL and is advertised by OpenAPI.
- ports must be between 1 and 65535; durations must be positive.
- `ALLOW_CREDENTIALS=true` cannot be combined with `ALLOW_ORIGINS=*`.
- auth target, service ID, and shared secret are required when auth is enabled.
- database pool limits must satisfy `0 <= DB_POOL_IDLE <= DB_POOL_MAX`.
- log level, format, outputs, and rotation values are validated before bootstrap.

The config loader creates a GORM/database pool without an eager network ping. This keeps process construction separate from dependency readiness; the first query or an explicit future readiness check establishes database connectivity.

## Authentication

The service uses `github.com/pecut-ai/auth-service/pkg/v2`. Set:

```dotenv
AUTH_ENABLED=true
AUTH_SERVICE_GRPC_TARGET=127.0.0.1:50051
SERVICE_ID=golang-clean-architecture
INTERNAL_SECRET=replace-me
```

Bearer tokens are verified by auth-service. The original local password/token implementation has been removed, so the template has one authentication authority. `GET /api/me` demonstrates reading the verified auth-service user context, and contacts are scoped directly by that external user ID.

The public process health endpoint is `GET /api/health`. All other example `/api` endpoints require bearer authentication. Setting `AUTH_ENABLED=false` is only useful for OpenAPI generation or bootstrap diagnostics; protected routes return `503 Service Unavailable` in that mode.

## API and OpenAPI

Huma is the runtime handler layer, not a parallel documentation-only router. Transport request/response types live in `internal/delivery/http/dto`, while handlers in `internal/delivery/http/route` authenticate, map DTOs into application models, call use cases, and map results. Request DTOs use Huma-native schema tags such as `format`, `minLength`, `maxLength`, `minimum`, and `default`.

With `DOCS_ENABLED=true`:

- interactive docs: `${APP_URL}/docs`
- OpenAPI JSON: `${APP_URL}/openapi.json`

Regenerate the checked-in contract with:

```bash
make openapi
```

Example endpoints:

- `GET /api/health`
- `GET /api/me`
- `GET|POST /api/contacts`
- `GET|PUT|DELETE /api/contacts/{contactId}`
- `GET|POST /api/contacts/{contactId}/addresses`
- `GET|PUT|DELETE /api/contacts/{contactId}/addresses/{addressId}`

## Logging

Logs are emitted through Zap. Production should use `LOG_FORMAT=json`; `console` is useful locally. `LOG_OUTPUT` may contain `stdout`, `stderr`, `file`, or a comma-separated combination. File output uses size/age/count rotation from the corresponding `LOG_ROTATION_*` settings.

Stable component examples are `api`, `gorm`, `repository`, `usecase`, `kafka_producer`, `kafka_consumer`, and `worker`. Auth-service log callbacks are bridged into the same logger as structured key/value fields.

Do not log bearer tokens, refresh tokens, passwords, or `INTERNAL_SECRET`.

## Build information

`internal/buildinfo` contains `Version`, `Commit`, and `BuildTime`, defaulting to development values. Release builds should stamp them with `-ldflags`; `make build VERSION=v1.2.3` does this automatically. The version appears in startup identity, Huma/OpenAPI metadata, and `/api/health`.

## Development

Requirements: Go 1.25.6+, PostgreSQL, the `migrate` CLI for migration commands, and an accessible auth-service instance. Kafka is only needed for producer or worker flows.

```bash
make install
make test
make run-web
```

Useful commands:

```bash
make build
make openapi
make migrate-up
make migrate-down
make run-worker
make clean
```

Focused tests intentionally avoid external services. Database/auth integration tests should be added beside the future multi-DB manager using explicit test infrastructure rather than package-level `init()` side effects.
