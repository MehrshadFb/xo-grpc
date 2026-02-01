package game

import "errors"

var (
	ErrInvalidCellIndex = errors.New("invalid cell index")
	ErrCellOccupied = errors.New("cell already occupied")
	ErrGameNotInProgress = errors.New("game is not in progress")
	ErrGameFinished = errors.New("game already finished")
	ErrInvalidPlayerMark = errors.New("invalid player mark")
	ErrNotPlayersTurn = errors.New("not player's turn")
)
