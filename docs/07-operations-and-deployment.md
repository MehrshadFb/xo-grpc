# Operations and Deployment

This document explains how the backend runs locally, how it is built, what it exposes, and how it fits into the deployed full-stack app.

## Runtime Configuration

Location: `internal/config/config.go`

Environment variables:

```text
PORT
  gRPC server port.
  Default: 50051

DATABASE_URL
  PostgreSQL connection string.
  Required by cmd/server.
```

Example:

```bash
DATABASE_URL='postgres://xo_user:xo_password@localhost:5432/xo_grpc?sslmode=disable' \
PORT=50051 \
go run ./cmd/server
```

## Local Development

Start Postgres:

```bash
docker compose up -d postgres
```

Run migrations:

```bash
make migrate-up
```

Run the server:

```bash
make run
```

Or with explicit environment:

```bash
DATABASE_URL='postgres://xo_user:xo_password@localhost:5432/xo_grpc?sslmode=disable' \
go run ./cmd/server
```

## Local Docker Compose

Repo file:

```text
docker-compose.yml
```

The repo compose is for local development. It starts:

- `xo-grpc`
- `postgres`

It exposes:

- `50051:50051` for gRPC
- `5432:5432` for local database access

This is convenient for development, but it is not the same as the public VPS production topology.

## Docker Image

Repo file:

```text
Dockerfile
```

The image uses a multi-stage build:

1. Builder stage:
   - Go Alpine image
   - installs `protobuf` and `make`
   - downloads Go modules
   - installs protobuf generators
   - runs `make proto`
   - builds `./cmd/server`

2. Runtime stage:
   - Alpine image
   - copies the compiled binary
   - exposes `50051`
   - runs `./xo-grpc`

Build command:

```bash
docker build -t xo-grpc .
```

## gRPC Reflection

The server enables gRPC reflection:

```go
reflection.Register(grpcServer)
```

This allows tools such as `grpcurl` to inspect and call services without manually passing proto files.

Example:

```bash
grpcurl -plaintext localhost:50051 xo.v1.HealthService/Health
```

## Metrics

The backend starts a separate HTTP metrics server:

```text
http://localhost:9090/metrics
```

Metrics:

```text
xo_grpc_requests_total
xo_grpc_request_duration_seconds
xo_active_watch_streams
```

The gRPC interceptors record request count and duration by:

- method
- status code
- request type (`unary` or `stream`)

## Logging

The server uses Go's `log/slog`.

Unary and streaming interceptors log:

- method
- gRPC status code
- duration
- stream type
- error when present

## Graceful Shutdown

The server listens for:

```text
SIGINT
SIGTERM
```

On shutdown:

1. gRPC server calls `GracefulStop`.
2. metrics HTTP server shuts down.
3. Postgres pool is closed by deferred cleanup.

## Production Topology

The current VPS deployment runs both frontend and backend on one server through Docker Compose under `/opt/xo`.

Public traffic:

```text
Internet
  |
  v
Caddy on 80/443
  |
  v
Next.js frontend
```

Private internal traffic:

```text
Next.js API routes
  |
  v
xo-grpc backend on Docker network
  |
  v
PostgreSQL on Docker network
```

Only Caddy is public. Backend gRPC and Postgres are not exposed publicly in the production compose.

Current public HTTPS hostname:

```text
https://xo-40-233-102-214.sslip.io
```

The plain IP HTTP URL also works:

```text
http://40.233.102.214
```

The `sslip.io` hostname maps the embedded IP address to the server IP. Caddy uses that hostname to request and renew a public TLS certificate automatically.

## Server Helper Script

Repo script:

```text
scripts/ssh-backend.sh
```

Purpose:

- SSH into the VPS without typing the full command each time.

Default behavior:

```text
host: 40.233.102.214
user: opc
key:  $HOME/.ssh/id_ed25519
```

Usage:

```bash
./scripts/ssh-backend.sh
```

Run one remote command:

```bash
./scripts/ssh-backend.sh 'cd /opt/xo && sudo docker compose ps'
```

The script does not contain private keys or passwords.

## Useful Production Commands

Check containers:

```bash
./scripts/ssh-backend.sh 'cd /opt/xo && sudo docker compose ps'
```

View logs:

```bash
./scripts/ssh-backend.sh 'cd /opt/xo && sudo docker compose logs --tail=100'
```

Follow logs:

```bash
./scripts/ssh-backend.sh 'cd /opt/xo && sudo docker compose logs -f --tail=100'
```

Restart stack:

```bash
./scripts/ssh-backend.sh 'cd /opt/xo && sudo docker compose up -d'
```

Check public health through the frontend:

```bash
curl https://xo-40-233-102-214.sslip.io/api/health
```

## Security Notes

Current public exposure:

- `80` and `443` to Caddy

Not public:

- backend gRPC `50051`
- Postgres `5432`
- raw Next.js `3000`
- metrics `9090`

The public browser API is the Next.js app. It calls the backend from server-side code.

The player session tokens are bearer tokens. Anyone with a token can act as that player for that game. For this project, that is acceptable and simple. For a larger product, consider:

- token expiration
- signed/encrypted tokens
- account authentication
- rate limits
- stricter logging redaction

