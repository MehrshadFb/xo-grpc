package session

import "errors"

var (
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionGameMismatch = errors.New("session game mismatch")
)
