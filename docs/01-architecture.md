# Architecture

The backend follows a layered architecture. Each layer has one clear responsibility and depends inward toward the application/domain code.

```text
Client / Frontend
      |
      v
gRPC transport layer
      |
      v
Application service layer
      |
      v
Domain model and repository interfaces
      |
      v
Repository implementation
      |
      v
PostgreSQL
```

## Layer Responsibilities

### API Contract

Location: `api/proto/xo/v1`

The `.proto` files define the public contract. This is what frontend/server clients should treat as the source of truth for request and response shapes.

The generated Go files are written to `gen/go`, but that directory is ignored by git. CI and Docker regenerate those files from the proto definitions.

### Entrypoint

Location: `cmd/server/main.go`

The server entrypoint wires everything together:

1. Load config from environment.
2. Register Prometheus metrics.
3. Create the realtime hub.
4. Connect to PostgreSQL.
5. Create repository implementations.
6. Create the session manager.
7. Create application services.
8. Create gRPC handlers.
9. Register gRPC services and reflection.
10. Start the gRPC server.
11. Start the metrics HTTP server.
12. Handle graceful shutdown on `SIGINT` and `SIGTERM`.

The entrypoint is intentionally the composition root. Most packages do not construct their own dependencies.

### Domain

Location: `internal/domain`

The domain layer is the business-rule core. It contains:

- game state
- players
- marks
- board state
- status transitions
- move validation
- win/draw detection
- rematch state
- room-level score

The domain does not import gRPC, PostgreSQL, HTTP, metrics, or config packages.

### Service Layer

Location: `internal/service`

The service layer coordinates use cases. It owns workflows such as:

- create a game
- join a game
- validate a session
- get game state
- make a move
- request a rematch
- publish realtime events

This layer uses domain methods to enforce rules, repository interfaces to persist state, and the realtime hub to notify watchers.

### Repository Boundary

Location: `internal/repository`

The repository package defines interfaces:

- `GameRepository`
- `SessionRepository`

Services depend on these interfaces instead of depending directly on PostgreSQL. That makes it easy to test services with the in-memory store and run production with the Postgres store.

### Store Implementations

Locations:

- `internal/store/postgres`
- `internal/store/memory`

The Postgres store is the production persistence implementation. It handles transactions, JSON board serialization, players, sessions, and optimistic locking.

The memory store is a test-friendly implementation. It stores games and sessions in maps protected by mutexes.

### Transport

Location: `internal/transport/grpc`

The transport layer adapts gRPC to the service layer:

- receives protobuf requests
- calls services
- maps service/domain errors to gRPC status codes
- converts domain objects to protobuf messages
- streams realtime events to clients
- records request metrics through interceptors

Transport code should stay thin. Business rules should live in domain/service packages.

## Dependency Direction

The main dependency rule is:

```text
transport -> service -> domain/repository interfaces
store     -> domain/repository interfaces
cmd       -> everything, for wiring only
```

The domain layer should remain independent. This keeps the game rules easy to test and easy to reason about.

## Runtime Components

The backend process exposes two servers:

```text
gRPC server
  default port: 50051
  services: LobbyService, GameService, HealthService

metrics HTTP server
  port: 9090
  endpoint: /metrics
```

The production deployment currently puts this backend behind the Next.js app. Browsers call the frontend, the frontend's server-side API routes call gRPC, and gRPC/Postgres stay private on the Docker network.

```text
Browser
  |
  v
Caddy HTTPS
  |
  v
Next.js frontend/API routes
  |
  v
xo-grpc backend
  |
  v
PostgreSQL
```

## Important Design Choices

### gRPC Instead of REST

gRPC gives this project:

- typed request/response contracts
- generated clients and servers
- efficient binary encoding
- server-side streaming for realtime game updates

The frontend still exposes normal HTTP routes to browsers through Next.js API routes, but internally it can use the typed gRPC contract.

### Full-State Events

Realtime events include the full `GameState`, not only a tiny patch. That makes clients simpler because they can replace local state with the newest server state.

### Optimistic Locking

Game updates use a `version` field. PostgreSQL updates only succeed when the row still has the expected previous version. This prevents silent overwrites when two requests race.

### In-Process Realtime Hub

The realtime hub is in memory. It is simple and fast for one backend instance. If the backend runs multiple replicas later, game events would need a shared pub/sub system such as Redis, NATS, or PostgreSQL notifications.

