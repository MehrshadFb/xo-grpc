# Testing and CI

The project uses several levels of tests:

```text
domain tests
service tests
repository tests
end-to-end gRPC tests
CI with PostgreSQL, migrations, race detector, and Docker build
```

## Test Commands

Run all tests:

```bash
go test ./...
```

Run with race detector:

```bash
go test -race ./...
```

Run Postgres integration tests:

```bash
DATABASE_URL='postgres://xo_user:xo_password@localhost:5432/xo_grpc?sslmode=disable' \
go test ./internal/store/postgres -v
```

Run e2e tests:

```bash
go test ./test/e2e -v
```

## Domain Tests

Location:

```text
internal/domain/game/rules_test.go
```

These tests exercise pure game behavior:

- initial game state
- start transition
- turn flow
- invalid cell indexes
- occupied cells
- wrong turn rejection
- game not in progress rejection
- X win
- draw
- no moves after finish
- rematch waits for both players
- rematch keeps score and resets board
- rematch requires finished game

These are the fastest and most important tests because they cover the core rules without infrastructure.

## Service Tests

Locations:

```text
internal/service/lobby/service_test.go
internal/service/game/service_test.go
internal/service/session/manager_test.go
```

These tests use the in-memory repository implementation.

Lobby service tests cover:

- create game
- join game
- invalid join code
- already full game
- empty display name

Game service tests cover:

- get state
- X moves first
- wrong turn rejection
- invalid token rejection
- O moves after X
- rematch requires both players
- rematch keeps score
- rematch before finish is rejected

Session manager tests cover:

- create and get token
- missing token
- token/game mismatch

## Memory Store Tests

Location:

```text
internal/store/memory/store_test.go
```

These tests check:

- create and get by id
- not found behavior
- duplicate game id
- duplicate join code
- get by join code
- update
- update missing game

The memory store is mostly a testing tool, but testing it protects service tests from false assumptions.

## Postgres Integration Tests

Locations:

```text
internal/store/postgres/game_repository_test.go
internal/store/postgres/session_repository_test.go
```

These tests require `DATABASE_URL`.

They verify:

- game create/get/update
- lookup by join code
- player persistence
- session persistence
- session survives game update
- board persistence
- status persistence
- optimistic version flow
- score/rematch persistence
- missing session behavior

The tests truncate tables before running so they can use fixed ids and join codes.

## End-to-End gRPC Tests

Locations:

```text
test/e2e/grpc_flow_test.go
test/e2e/watch_resume_test.go
```

These tests start an actual in-process gRPC server with:

- memory game store
- memory session store
- realtime hub
- lobby handler
- game handler

They then call generated gRPC clients.

Covered flows:

- create game
- open WatchGame stream
- receive initial snapshot
- join game
- receive player joined event
- make move
- receive move event
- finish game
- receive game over event
- request rematch
- receive rematch requested event
- second rematch request starts next round
- receive round started event
- reconnect stream with `after_version`
- skip duplicate snapshot on resume
- receive future events after resume

These tests are valuable because they validate the transport, service, domain, and realtime layers together.

## GitHub Actions CI

Location:

```text
.github/workflows/test.yml
```

CI runs on:

- push to `main`
- pull requests

CI steps:

1. Checkout repository.
2. Setup Go using `go.mod`.
3. Install `protoc`.
4. Install Go protobuf plugins.
5. Download dependencies.
6. Generate protobuf code with `make proto`.
7. Start PostgreSQL service.
8. Install `golang-migrate`.
9. Run migrations.
10. Run `go test -race ./...`.
11. Run Postgres integration tests.
12. Build Docker image.

This gives confidence that:

- proto generation works from clean checkout
- migrations work
- tests pass with race detector
- Postgres integration is healthy
- Docker build is not broken

## What the Current Tests Do Not Cover Yet

Useful future additions:

- gRPC handler error-code tests
- direct tests for logging/metrics interceptors
- Postgres optimistic-lock conflict test
- MakeMove retry or conflict behavior, if added later
- load test for many WatchGame subscribers
- deployment smoke test outside the backend repo

For the current project size, the existing coverage is strong and well aligned with the main risks.

