package server

import "errors"

var (
	ErrHandlerNotFound = errors.New("Handler for that command could not be found")
	ErrPlayerCantJoin  = errors.New("Player is unable to join")
	ErrUnexpectedCom   = errors.New("Unexpected communication")
	ErrEmptyBuffer     = errors.New("Buffer is empty")

	ErrClientRejected     = errors.New("Client was rejected by the server")
	ErrClientNotConnected = errors.New("Client is not connected to a server")
)
