package server

import "errors"

var (
	ErrHandlerNotFound = errors.New("Handler for that command could not be found")
	ErrPlayerCantJoin  = errors.New("Player is unable to join")
)
