package server

import "errors"

var (
	ErrHandlerNotFound = errors.New("Handler for that command could not be found")
	ErrPlayerCantJoin  = errors.New("Player is unable to join")
	ErrUnexpectedCom   = errors.New("Unexpected communication")

	ErrClientRejected = errors.New("Client was rejected by the server")
)
