# Persistence and Migrations

Persistence is split behind repository interfaces.

```text
internal/repository
  Interfaces used by services.

internal/store/postgres
  Production implementation.

internal/store/memory
  Test/in-memory implementation.
```

## Repository Interfaces

Location: `internal/repository`

### GameRepository

```go
type GameRepository interface {
    Create(g *game.Game) error
    GetByID(id string) (*game.Game, error)
    GetByJoinCode(joinCode string) (*game.Game, error)
    Update(g *game.Game) error
}
```

### SessionRepository

```go
type SessionRepository interface {
    Create(s session.Session) error
    Get(token string) (session.Session, error)
}
```

Services depend on these interfaces, not on Postgres directly.

## PostgreSQL Store

Location: `internal/store/postgres`

The Postgres store is used by the real server. It uses `pgxpool` for database access.

### Game Creation

`GameRepository.Create`:

1. serializes the board to JSON
2. starts a transaction
3. inserts into `games`
4. inserts Player X and Player O when present
5. commits the transaction

Creation uses a transaction because a game and its players should be written together.

### Loading a Game

Games can be loaded by:

- id
- join code

Loading does:

1. query `games`
2. convert string fields back to domain enums
3. convert JSON board back to `[9]Mark`
4. query players for the game
5. assign Player X and Player O on the domain object

### Updating a Game

`GameRepository.Update` writes the mutable game fields:

- status
- board
- next turn
- winner
- draw flag
- move number
- version
- score
- round number
- rematch flags
- updated timestamp

It also upserts players.

## Optimistic Locking

The Postgres update uses this idea:

```text
expectedVersion = game.Version - 1

UPDATE games
SET version = game.Version, ...
WHERE id = game.ID AND version = expectedVersion
```

If no rows are updated, another request already changed the game. The repository returns:

```go
repository.ErrConflict
```

This prevents a stale copy of the game from silently overwriting a newer copy.

Current behavior:

- rematch requests retry conflicts up to 3 times
- regular moves return the conflict as an internal error path through transport if it happens

For normal gameplay, turn validation makes conflicting moves less likely, but concurrent duplicate requests are still possible.

## Board Storage

The domain board is `[9]Mark`. PostgreSQL stores it as JSONB.

Example stored shape:

```json
["EMPTY","X","EMPTY","O","EMPTY","EMPTY","EMPTY","EMPTY","EMPTY"]
```

Mapper functions convert:

```text
domain Mark <-> database string
domain GameStatus <-> database string
[9]Mark <-> JSONB array
```

## Schema

Migrations live in:

```text
migrations
```

### `001_create_games`

Creates:

- `games`
- `players`
- `sessions`

#### `games`

Important columns:

- `id`
- `join_code`
- `status`
- `board`
- `next_turn`
- `winner`
- `is_draw`
- `move_number`
- `version`
- timestamps

#### `players`

Important columns:

- `id`
- `game_id`
- `display_name`
- `mark`

Constraint:

```text
UNIQUE(game_id, mark)
```

This prevents a game from having two X players or two O players.

#### `sessions`

Important columns:

- `token`
- `game_id`
- `player_id`
- `mark`

Sessions reference both game and player rows.

### `002_add_match_score_and_rematch`

Adds room-level match state:

- `x_wins`
- `o_wins`
- `draws`
- `round_number`
- `rematch_x_requested`
- `rematch_o_requested`

These fields make rematches possible without creating a new room.

## In-Memory Store

Location: `internal/store/memory`

The memory store uses maps:

```text
byID: game id -> game
byJoinCode: join code -> game id
```

It protects access with `sync.RWMutex`.

It clones games on read/write so tests do not accidentally mutate store state without calling `Update`.

The memory store is used in:

- domain-adjacent service tests
- e2e tests that start an in-process gRPC server

## Database Connection

Location: `internal/database/postgres.go`

`NewPostgresPool` creates a `pgxpool.Pool` with retry:

- max attempts: 10
- retry delay: 1 second

This is useful in Docker/CI where the app may start while Postgres is still becoming ready.

## Migration Commands

The Makefile wraps `golang-migrate`:

```bash
make migrate-up
make migrate-down
make migrate-force VERSION=<version>
```

Default local `DATABASE_URL`:

```text
postgres://xo_user:xo_password@localhost:5432/xo_grpc?sslmode=disable
```

