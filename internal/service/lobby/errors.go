package lobby

import "errors"

var (
	ErrEmptyDisplayName = errors.New("display name is required")
	ErrGameFull         = errors.New("game is full")
	ErrGameNotWaiting   = errors.New("game is not waiting")
	ErrInvalidJoinCode  = errors.New("invalid join code")
)