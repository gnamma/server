package server

import "errors"

var (
	ErrHandlerNotFound    = errors.New("Handler for that command could not be found")
	ErrPlayerCantJoin     = errors.New("Player is unable to join")
	ErrPlayerDoesntExist  = errors.New("Player does not exist")
	ErrNodeDoesntExist    = errors.New("Node does not exist")
	ErrNodeAlreadyExists  = errors.New("Node already exists")
	ErrUnexpectedCom      = errors.New("Unexpected communication")
	ErrEmptyBuffer        = errors.New("Buffer is empty")
	ErrClientDisconnected = errors.New("Client is disconnected")

	ErrClientRejected     = errors.New("Client was rejected by the server")
	ErrClientNotConnected = errors.New("Client is not connected to a server")
)
