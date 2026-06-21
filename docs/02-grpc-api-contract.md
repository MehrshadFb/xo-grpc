# gRPC API Contract

The public API is defined in `api/proto/xo/v1`.

The proto files are the contract between backend and clients. In this project, the Next.js frontend does not call gRPC directly from the browser. Instead:

```text
Browser
  |
  v
Next.js API route
  |
  v
gRPC client
  |
  v
xo-grpc backend
```

This keeps browser-facing code HTTP-friendly while preserving a strongly typed backend contract.

## Proto Files

```text
common.proto
  Shared enums and messages.

lobby.proto
  Create and join game room operations.

game.proto
  Game reads, moves, rematches, and streaming.

health.proto
  Health/readiness check.
```

## Shared Types

Location: `api/proto/xo/v1/common.proto`

### `Mark`

Represents a board cell or player mark.

```text
MARK_UNSPECIFIED
MARK_EMPTY
MARK_X
MARK_O
```

Domain equivalent: `internal/domain/game.Mark`

### `GameStatus`

Represents the game lifecycle.

```text
GAME_STATUS_WAITING
  Host created the room; waiting for opponent.

GAME_STATUS_IN_PROGRESS
  Both players joined; moves are allowed.

GAME_STATUS_FINISHED
  A win or draw happened.

GAME_STATUS_ABORTED
  Reserved for future resign/timeout behavior.
```

Domain equivalent: `internal/domain/game.GameStatus`

### `GameState`

This is the main state object returned by most API methods and streamed in events.

Fields include:

- `game_id`
- `join_code`
- `status`
- `board`
- `player_x`
- `player_o`
- `next_turn`
- `winner`
- `is_draw`
- `version`
- `score`
- `rematch`
- `round_number`

The board is always nine cells in row-major order:

```text
0 | 1 | 2
3 | 4 | 5
6 | 7 | 8
```

## LobbyService

Location: `api/proto/xo/v1/lobby.proto`

### `CreateGame`

Creates a new game room.

Request:

```proto
message CreateGameRequest {
  string display_name = 1;
}
```

Response:

```proto
message CreateGameResponse {
  GameState state = 1;
  string player_token = 2;
}
```

Behavior:

1. Creates a new game.
2. Assigns the creator as Player X.
3. Generates a join code.
4. Creates a player session token.
5. Returns waiting game state and token.

The returned token proves the caller is Player X for that game.

### `JoinGame`

Joins an existing waiting game by join code.

Request:

```proto
message JoinGameRequest {
  string join_code = 1;
  string display_name = 2;
}
```

Response:

```proto
message JoinGameResponse {
  GameState state = 1;
  string player_token = 2;
}
```

Behavior:

1. Finds the game by join code.
2. Validates that the game is waiting.
3. Assigns the caller as Player O.
4. Starts the game.
5. Creates a player session token.
6. Publishes a `PLAYER_JOINED` realtime event.

The returned token proves the caller is Player O for that game.

## GameService

Location: `api/proto/xo/v1/game.proto`

### `GetState`

Fetches the current game state.

Request:

```proto
message GetStateRequest {
  string game_id = 1;
  string player_token = 2;
}
```

The token must belong to the requested game. Otherwise the backend returns `PERMISSION_DENIED`.

### `MakeMove`

Applies one move for the authenticated player.

Request:

```proto
message MakeMoveRequest {
  string game_id = 1;
  string player_token = 2;
  int32 cell_index = 3;
}
```

Behavior:

1. Validates token and game id.
2. Loads the latest game.
3. Uses the session mark to determine whether the caller is X or O.
4. Calls domain `ApplyMove`.
5. Persists the new game state.
6. Publishes either `MOVE_MADE` or `GAME_OVER`.

Clients do not send their mark. The mark comes from the session token.

### `RequestRematch`

Requests another round in the same room after the game is finished.

Request:

```proto
message RequestRematchRequest {
  string game_id = 1;
  string player_token = 2;
}
```

Behavior:

1. Validates token and game id.
2. Requires the game to be finished.
3. Marks the player's rematch request.
4. If only one player has requested, the game stays finished.
5. If both players have requested, the round resets and score remains.
6. Publishes `REMATCH_REQUESTED` or `ROUND_STARTED`.

### `WatchGame`

Server-side streaming endpoint for realtime updates.

Request:

```proto
message WatchGameRequest {
  string game_id = 1;
  string player_token = 2;
  int64 after_version = 3;
}
```

Response stream:

```proto
message GameEvent {
  GameEventType type = 1;
  int64 event_id = 2;
  GameState state = 3;
  Move move = 4;
  GameOverReason game_over_reason = 5;
  string message = 6;
}
```

Resume behavior:

- If current game version is greater than `after_version`, the server first sends a `STATE_SNAPSHOT`.
- If the client already has the latest version, the server skips the snapshot and waits for future events.

This avoids duplicate snapshots after reconnects.

## HealthService

Location: `api/proto/xo/v1/health.proto`

### `Health`

Checks whether the service is ready.

If PostgreSQL is configured, the health service pings the database. It returns:

```text
HEALTH_STATUS_SERVING
HEALTH_STATUS_NOT_SERVING
```

## Error Mapping

The transport layer converts internal errors into gRPC status codes.

Common mappings:

```text
Invalid input
  -> INVALID_ARGUMENT

Missing/unknown session token
  -> UNAUTHENTICATED

Token belongs to a different game
  -> PERMISSION_DENIED

Game not found
  -> NOT_FOUND

Wrong turn, occupied cell, finished game, not waiting, full game
  -> FAILED_PRECONDITION

Unexpected repository/database/internal failure
  -> INTERNAL
```

This keeps clients from depending on Go error types.

## Generated Code

Generation command:

```bash
make proto
```

Generated files go to:

```text
gen/go/xo/v1
```

That folder is ignored by git. The source of truth is the `.proto` files.

