package server

import (
	"encoding/json"
	"io"
	"log"
)

type CommunicationHandler func(buf io.Reader, conn Conn) error

type Room struct {
	s *Server

	handlers map[string]CommunicationHandler // Do not change at runtime.

	players     map[uint]*Player
	playerCount uint
}

func NewRoom(s *Server) *Room {
	r := &Room{
		s:       s,
		players: make(map[uint]*Player),
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

func (r *Room) CanJoin(p *Player) bool {
	_, ok := r.players[p.ID]

	return p.Valid() && !ok
}

func (r *Room) connectRequest(buf io.Reader, conn Conn) error {
	c := ConnectRequest{}

	err := json.NewDecoder(buf).Decode(&c)
	if err != nil {
		return err
	}

	cv := ConnectVerdict{
		CanProceed: true,
		Message:    "Welcome to the server!",
	}

	p, err := r.Join(c.Username)
	if err != nil {
		cv = ConnectVerdict{
			CanProceed: false,
			Message:    "Sorry. Connection rejected.",
		}
	}

	log.Println("Connected player:", p)

	return conn.Send(ConnectVerdictCmd, &cv)
}

func (r *Room) Join(u string) (*Player, error) {
	p := &Player{
		Username: u,
		ID:       r.playerCount + 1, // Don't increment straight away so that to prevent an overflow.
	}

	if !r.CanJoin(p) {
		return nil, ErrPlayerCantJoin
	}

	r.playerCount += 1

	return p, nil
}

func (r *Room) ping(buf io.Reader, conn Conn) error {
	defer conn.Close()

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
