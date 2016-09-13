package server

import (
	"encoding/json"
	"io"
	"net"
	"time"
)

type CommunicationHandler func(buf io.Reader, conn net.Conn) error

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

func (r *Room) Respond(cmd string, buf io.Reader, conn net.Conn) error {
	f, ok := r.handlers[cmd]
	if !ok {
		return ErrHandlerNotFound
	}

	return f(buf, conn)
}

func (r *Room) CanJoin(p Player) bool {
	return p.Valid()
}

func (r *Room) connectRequest(buf io.Reader, conn net.Conn) error {
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

	cv.Communication = Communication{
		Command: ConnectVerdictCmd,
		SentAt:  time.Now().UnixNano(),
	}

	out, err := json.Marshal(cv)
	if err != nil {
		return err
	}

	conn.Write(out)

	return nil
}

func (r *Room) ping(buf io.Reader, conn net.Conn) error {
	pi := Ping{}

	err := json.NewDecoder(buf).Decode(&pi)
	if err != nil {
		return err
	}

	po := Pong{
		Communication: Communication{
			Command: PongCmd,
			SentAt:  time.Now().UnixNano(),
		},
		ReceivedAt: pi.SentAt,
	}

	out, err := json.Marshal(po)
	if err != nil {
		return err
	}

	conn.Write(out)

	return nil
}
