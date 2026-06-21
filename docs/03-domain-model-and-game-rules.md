# Domain Model and Game Rules

The domain layer is the core of the project. It lives under:

```text
internal/domain/game
internal/domain/session
```

This layer should be easy to test without a database or a gRPC server.

## Game Type

Location: `internal/domain/game/types.go`

The main domain object is `Game`:

```go
type Game struct {
    ID                string
    JoinCode          string
    Status            GameStatus
    Board             [9]Mark
    NextTurn          Mark
    Winner            Mark
    IsDraw            bool
    MoveNumber        int64
    Version           int64
    PlayerX           *Player
    PlayerO           *Player
    XWins             int64
    OWins             int64
    Draws             int64
    RoundNumber       int64
    RematchXRequested bool
    RematchORequested bool
}
```

The domain object contains both current-round state and room-level match state.

Current-round state:

- board
- next turn
- winner
- draw flag
- move number
- status

Room-level state:

- players
- score
- round number
- rematch request flags

## Board

The board is a fixed array:

```go
[9]Mark
```

Indexes are row-major:

```text
0 | 1 | 2
3 | 4 | 5
6 | 7 | 8
```

Using `[9]Mark` instead of `[]Mark` gives a fixed-size board at the domain level. The protobuf representation is still a repeated field for wire compatibility.

## Marks

Domain marks:

```text
MarkUnspecified
MarkEmpty
MarkX
MarkO
```

Only `MarkX` and `MarkO` can play moves. `MarkEmpty` represents an empty board cell and no winner.

## Game Status

Domain statuses:

```text
StatusWaiting
  Created and waiting for Player O.

StatusInProgress
  Both players joined and moves are allowed.

StatusFinished
  Win or draw.

StatusAborted
  Reserved for future timeout/resign behavior.
```

## Creating a Game

Location: `internal/domain/game/new.go`

`NewGame(id, joinCode)` initializes:

- `StatusWaiting`
- `NextTurn = MarkX`
- `Version = 1`
- `RoundNumber = 1`
- every board cell as `MarkEmpty`

It does not assign players. Player assignment is done by the service layer through `SetPlayerX` and `SetPlayerO`.

## Starting a Game

Location: `internal/domain/game/rules.go`

`Start()` moves a game from waiting to in progress.

Requirements:

- game must be waiting
- Player O must exist
- game must not be finished

Effects:

- status becomes `StatusInProgress`
- next turn is `MarkX`
- version increments

Calling `Start()` on a game that is already in progress is treated as a no-op.

## Applying a Move

Location: `internal/domain/game/rules.go`

`ApplyMove(player, cell)` is the main rule method.

Validation order:

1. Cell index must be `0..8`.
2. Player mark must be X or O.
3. Game must not be finished.
4. Game must be in progress.
5. It must be that player's turn.
6. Cell must be empty.

Effects:

1. Places the mark on the board.
2. Increments `MoveNumber`.
3. Increments `Version`.
4. Checks for a winner.
5. Checks for a draw.
6. If not finished, switches turn.

If the move wins:

- status becomes `StatusFinished`
- winner becomes X or O
- score increments for the winner

If the move draws:

- status becomes `StatusFinished`
- winner becomes `MarkEmpty`
- `IsDraw` becomes true
- draw score increments

## Win Detection

The winner check uses the eight normal Tic-Tac-Toe lines:

```text
Rows:
0 1 2
3 4 5
6 7 8

Columns:
0 3 6
1 4 7
2 5 8

Diagonals:
0 4 8
2 4 6
```

If all three cells in a line are equal and non-empty, that mark wins.

## Draw Detection

A draw requires:

- every cell is filled
- there is no winner

The code checks for any empty cell first, then confirms there is no winner.

## Score

The score belongs to the room, not only one round:

- `XWins`
- `OWins`
- `Draws`

`recordOutcome()` increments score exactly when a round finishes.

The score persists across rematches.

## Rematch

Location: `internal/domain/game/rules.go`

`RequestRematch(player)` handles replaying in the same room.

Requirements:

- caller mark must be X or O
- game must be finished
- Player O must exist

First request:

- sets that player's rematch flag
- increments version
- leaves status as `StatusFinished`
- returns `RematchRequested`

Duplicate request by the same player:

- no state change
- no version bump
- returns `RematchNoop`

Second player request:

- resets the board
- sets status to `StatusInProgress`
- resets move number
- resets winner/draw state
- increments round number
- clears rematch flags
- keeps the score
- returns `RematchStarted`

Current behavior: X always starts every new round.

## Version

`Version` is a monotonic state version used by:

- streaming resume logic
- optimistic locking in PostgreSQL
- event ids

A state-changing operation should bump the version. No-op rematch requests intentionally do not.

## Domain Errors

Location: `internal/domain/game/errors.go`

Important domain errors:

```text
ErrInvalidCellIndex
ErrCellOccupied
ErrGameNotInProgress
ErrGameFinished
ErrInvalidPlayerMark
ErrNotPlayersTurn
ErrPlayerOMissing
ErrGameNotFinished
```

The transport layer maps these to gRPC status codes.

## Session Domain

Location: `internal/domain/session/session.go`

`Session` ties a token to a player and game:

```go
type Session struct {
    Token    string
    GameID   string
    PlayerID string
    Mark     game.Mark
}
```

The key idea: clients do not choose their mark when making a move. The backend derives the mark from the session token.

