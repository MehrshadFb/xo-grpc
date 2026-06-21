# Service Layer Workflows

The service layer lives under:

```text
internal/service
```

It turns domain rules into application workflows. It knows about repositories, sessions, and realtime events.

## Service Packages

```text
internal/service/lobby
  Create and join games.

internal/service/game
  Read state, make moves, request rematches.

internal/service/session
  Create and validate player sessions.

internal/service/health
  Check database readiness.
```

## Lobby Service

Location: `internal/service/lobby/service.go`

The lobby service owns room creation and room joining.

### CreateGame Flow

```text
CreateGame(displayName)
  |
  |-- trim and validate display name
  |-- generate game id
  |-- generate join code
  |-- generate player id
  |-- create domain game
  |-- set creator as Player X
  |-- persist game
  |-- create session token for Player X
  v
return game state + player token
```

Important behavior:

- creator is always X
- newly created game starts in `StatusWaiting`
- returned token is required for future game calls

### JoinGame Flow

```text
JoinGame(joinCode, displayName)
  |
  |-- trim and validate join code/display name
  |-- load game by join code
  |-- require game to be waiting
  |-- require Player O to be empty
  |-- generate Player O id
  |-- set Player O
  |-- start game
  |-- persist game
  |-- publish PlayerJoined event
  |-- create session token for Player O
  v
return game state + player token
```

Important behavior:

- second player is always O
- joining starts the game
- joining publishes a realtime event so waiting clients update immediately

## ID Generation

Location: `internal/service/lobby/ids.go`

Generated IDs:

- game id: 16 random bytes encoded as hex
- player id: 16 random bytes encoded as hex
- join code: 6 characters from `A-Z0-9`

The code uses `crypto/rand`, not pseudo-random math/rand.

## Session Manager

Location: `internal/service/session/manager.go`

The session manager creates and validates player tokens.

### Create

`Create(gameID, playerID, mark)`:

1. validates game id and player id
2. validates mark is X or O
3. generates a 24-byte random token as hex
4. stores the session through `SessionRepository`
5. returns the token

### ValidateGame

`ValidateGame(token, gameID)`:

1. loads the session by token
2. checks that `session.GameID == gameID`
3. returns the session if valid

This prevents a valid token from one game being used in another game.

## Game Service

Location: `internal/service/game/service.go`

The game service owns gameplay operations after a room exists.

### GetState Flow

```text
GetState(gameID, token)
  |
  |-- trim and validate input
  |-- validate token belongs to game
  |-- load game by id
  v
return game
```

Even read operations require a player token. That means only room participants can read the game through the gRPC API.

### MakeMove Flow

```text
MakeMove(gameID, token, cell)
  |
  |-- trim and validate input
  |-- validate token belongs to game
  |-- use session mark as caller mark
  |-- load current game
  |-- call domain ApplyMove(mark, cell)
  |-- persist updated game
  |-- publish MoveMade or GameOver event
  v
return game
```

The service does not trust the client to say whether it is X or O. It derives the mark from the persisted session.

If the move finishes the round:

- event type is `GameOver`
- game over reason is `win` or `draw`

Otherwise:

- event type is `MoveMade`

### RequestRematch Flow

```text
RequestRematch(gameID, token)
  |
  |-- trim and validate input
  |-- validate token belongs to game
  |-- try up to 3 times:
  |     |-- load game
  |     |-- call domain RequestRematch(mark)
  |     |-- persist updated game
  |     |-- retry if optimistic-lock conflict
  |-- publish RematchRequested or RoundStarted event if changed
  v
return game
```

The retry loop exists because both players may request rematch at nearly the same time. If the repository detects a version conflict, the service reloads and tries again.

The result tracks:

- `Changed`: whether state changed
- `Started`: whether the second request started the next round

Duplicate rematch requests are idempotent and do not publish another event.

## Health Service

Location: `internal/service/health/service.go`

`Ready(ctx)` checks if PostgreSQL is reachable by pinging the pool.

If no database pool is configured, it returns ready. In the current server, Postgres is required before services are created, so production readiness depends on DB connectivity.

## Error Ownership

Each layer owns its own errors:

```text
domain/game
  rule errors such as invalid move, wrong turn, game finished

service/lobby
  workflow errors such as empty display name, invalid join code, game full

service/session
  session errors such as not found or game mismatch

repository
  persistence boundary errors such as optimistic conflict
```

The gRPC transport maps these errors to public gRPC status codes.

