package memory

import "errors"

var (
	ErrGameNotFound = errors.New("game not found")
	ErrJoinCodeNotFound = errors.New("join code not found")
	ErrDuplicateGameID = errors.New("duplicate game id")
	ErrDuplicateJoinCode = errors.New("duplicate join code")
)
