package server

import (
	"encoding/json"
	"io"
)

type CommunicationHandler func(buf io.Reader, conn Conn) error

type Room struct {
	s *Server

	handlers map[string]CommunicationHandler // Do not change at runtime.
}

func NewRoom(s *Server) *Room {
	r := &Room{
		s: s,
	}

	r.handlers = map[string]CommunicationHandler{
		PingCmd:           r.ping,
		ConnectRequestCmd: r.connectRequest,
	}

	return r
}

func (r *Room) Respond(cmd string, buf io.Reader, conn Conn) error {
	f, ok := r.handlers[cmd]
	if !ok {
		return ErrHandlerNotFound
	}

	return f(buf, conn)
}

func (r *Room) CanJoin(p Player) bool {
	return p.Valid()
}

func (r *Room) connectRequest(buf io.Reader, conn Conn) error {
	c := ConnectRequest{}

	err := json.NewDecoder(buf).Decode(&c)
	if err != nil {
		return err
	}

	p := Player{
		Username: c.Username,
	}

	cv := ConnectVerdict{}

	if r.s.Room.CanJoin(p) {
		cv = ConnectVerdict{
			CanProceed: true,
			Message:    "Welcome to the server!",
		}
	} else {
		cv = ConnectVerdict{
			CanProceed: false,
			Message:    "Sorry. Connection rejected.",
		}
	}

	return conn.Send(ConnectVerdictCmd, &cv)
}

func (r *Room) ping(buf io.Reader, conn Conn) error {
	pi := Ping{}

	err := json.NewDecoder(buf).Decode(&pi)
	if err != nil {
		return err
	}

	po := Pong{
		ReceivedAt: pi.SentAt,
	}

	return conn.Send(PongCmd, &po)
}
