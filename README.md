# xo-grpc

A multiplayer Tic-Tac-Toe backend built with Go, gRPC, Protocol Buffers, PostgreSQL, and server-side streaming.

## Features

- Create and join games
- Turn-based gameplay with validation
- Real-time game updates using gRPC server-side streaming
- PostgreSQL persistence
- Session-based player authentication
- Optimistic locking for concurrent updates
- Reconnect/resume support for game streams
- Health and readiness checks
- Prometheus metrics
- Structured logging
- Graceful shutdown
- Docker and Docker Compose support
- GitHub Actions CI
- Unit, integration, race, and end-to-end tests

---

## Architecture

```text
Client
   |
   v
gRPC API
   |
   v
Service Layer
   |
   v
Repository Layer
   |
   v
PostgreSQL
```

### Project Structure

```text
api/proto/xo/v1
├── Protocol Buffer definitions

cmd/server
├── Application entrypoint

internal/config
├── Configuration loading

internal/database
├── PostgreSQL connection management

internal/domain
├── Core business entities

internal/repository
├── Repository interfaces

internal/service
├── Business logic

internal/store/postgres
├── PostgreSQL repository implementations

internal/realtime
├── Realtime event hub

internal/metrics
├── Prometheus metrics

internal/transport/grpc
├── gRPC handlers and interceptors

migrations
├── Database schema migrations

test/e2e
├── End-to-end tests
```

---

## Technologies

- Go
- gRPC
- Protocol Buffers
- PostgreSQL
- pgx
- Prometheus
- Docker
- Docker Compose
- GitHub Actions

---

## Requirements

- Go 1.24+
- Docker
- Docker Compose
- protoc
- grpcurl

---

## Installation

```bash
git clone https://github.com/MehrshadFb/xo-grpc.git
cd xo-grpc
```

Install protobuf generators:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

## Environment Variables


| Variable     | Description                  |
| ------------ | ---------------------------- |
| DATABASE_URL | PostgreSQL connection string |
| PORT         | gRPC server port             |


Example:

```bash
DATABASE_URL=postgres://xo_user:xo_password@localhost:5432/xo_grpc?sslmode=disable
PORT=50051
```

---

## Running Locally

Start PostgreSQL:

```bash
docker compose up -d postgres
```

Run migrations:

```bash
make migrate-up
```

Start the server:

```bash
DATABASE_URL='postgres://xo_user:xo_password@localhost:5432/xo_grpc?sslmode=disable' \
go run ./cmd/server
```

---

## Docker

Build and run:

```bash
docker compose up --build
```

Stop:

```bash
docker compose down
```

---

## Protobuf Generation

```bash
make proto
```

---

## Database Migrations

Apply migrations:

```bash
make migrate-up
```

Rollback migrations:

```bash
make migrate-down
```

---

## gRPC Services

### LobbyService

```text
CreateGame
JoinGame
```

### GameService

```text
GetState
MakeMove
WatchGame
```

### HealthService

```text
Health
```

---

## Realtime Streaming

`WatchGame` uses server-side streaming.

Clients can reconnect using:

```proto
message WatchGameRequest {
  string game_id = 1;
  string player_token = 2;
  int64 after_version = 3;
}
```

If the client already has the latest version, the server skips sending a duplicate snapshot and waits for future updates.

---

## Health Checks

```bash
grpcurl -plaintext localhost:50051 xo.v1.HealthService/Health
```

Example response:

```json
{
  "status": "HEALTH_STATUS_SERVING",
  "message": "ok"
}
```

---

## Metrics

Prometheus metrics are exposed on:

```text
http://localhost:9090/metrics
```

Example:

```bash
curl localhost:9090/metrics
```

Available metrics:

```text
xo_grpc_requests_total
xo_grpc_request_duration_seconds
xo_active_watch_streams
```

---

## Testing

Run all tests:

```bash
go test ./...
```

Run race detector:

```bash
go test -race ./...
```

Run PostgreSQL integration tests:

```bash
DATABASE_URL='postgres://xo_user:xo_password@localhost:5432/xo_grpc?sslmode=disable' \
go test ./internal/store/postgres -v
```

Run end-to-end tests:

```bash
go test ./test/e2e -v
```

---

## CI Pipeline

GitHub Actions automatically performs:

- Dependency installation
- Protobuf generation
- Migration execution
- Unit tests
- Integration tests
- Race tests
- Docker image build

---

## Concurrency Control

Game updates use optimistic locking.

Each game maintains a version number.

Updates succeed only when the expected version matches the current database version, preventing stale writes from overwriting newer state.

---

## Observability

- Structured logging using `slog`
- Prometheus metrics
- Health checks
- Readiness checks

---

## License

MIT