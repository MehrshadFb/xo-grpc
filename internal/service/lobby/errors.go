package lobby

import "errors"

var (
	ErrGameFull        = errors.New("game is full")
	ErrGameNotWaiting  = errors.New("game is not waiting")
	ErrInvalidJoinCode = errors.New("invalid join code")
)