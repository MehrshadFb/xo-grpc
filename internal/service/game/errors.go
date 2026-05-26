package gamesvc

import "errors"

var (
	ErrInvalidGameID = errors.New("invalid game id")
	ErrInvalidToken  = errors.New("invalid player token")
)
