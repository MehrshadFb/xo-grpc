# xo-grpc Documentation

This folder is a learning-oriented reference for the backend. It explains how the project is structured, why the layers exist, and how requests move through Go, gRPC, PostgreSQL, realtime streaming, tests, and deployment.

The backend is a multiplayer Tic-Tac-Toe service. It is intentionally small enough to understand end to end, but it uses production-style patterns:

- gRPC and Protocol Buffers for the API contract
- Go packages split by responsibility
- a pure domain layer for game rules
- services for application workflows
- repository interfaces for persistence boundaries
- PostgreSQL storage with migrations
- server-side streaming for realtime updates
- session tokens for player identity
- Prometheus metrics and structured logs
- Docker and GitHub Actions CI

## Reading Order

1. [Architecture](./01-architecture.md)
2. [gRPC API Contract](./02-grpc-api-contract.md)
3. [Domain Model and Game Rules](./03-domain-model-and-game-rules.md)
4. [Service Layer Workflows](./04-service-layer-workflows.md)
5. [Persistence and Migrations](./05-persistence-and-migrations.md)
6. [Realtime Streaming](./06-realtime-streaming.md)
7. [Operations and Deployment](./07-operations-and-deployment.md)
8. [Testing and CI](./08-testing-and-ci.md)

## Mental Model

At a high level, the backend is a layered application:

```text
gRPC client
   |
   v
transport/grpc handlers
   |
   v
service layer
   |
   v
domain model + repository interfaces
   |
   v
PostgreSQL repository implementation
```

The important design idea is that the game rules do not know about gRPC, PostgreSQL, Docker, or metrics. The domain package only knows how a game should behave. Everything else is an adapter around that core.

## Main Packages

```text
api/proto/xo/v1
  Public Protocol Buffer contract.

cmd/server
  Application entrypoint and dependency wiring.

internal/domain
  Core game and session types. This is the business-rule center.

internal/service
  Use-case workflows: create room, join room, make move, request rematch.

internal/repository
  Interfaces that service code depends on.

internal/store/postgres
  PostgreSQL implementation of repository interfaces.

internal/store/memory
  In-memory implementation used mainly by unit and e2e tests.

internal/realtime
  In-process pub/sub hub for game update streaming.

internal/transport/grpc
  gRPC handlers, error mapping, protobuf mapping, logging interceptors.

internal/config
  Environment configuration.

internal/database
  PostgreSQL connection pool with retry.

internal/metrics
  Prometheus metrics registration.

migrations
  Database schema migrations.

test/e2e
  End-to-end gRPC tests.
```

