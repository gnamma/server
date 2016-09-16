package server

import "log"

type CommunicationHandler func(conn Conn) error

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

func (r *Room) Respond(cmd string, conn Conn) error {
	f, ok := r.handlers[cmd]
	if !ok {
		return ErrHandlerNotFound
	}

	return f(conn)
}

func (r *Room) CanJoin(p *Player) bool {
	_, ok := r.players[p.ID]

	return p.Valid() && !ok
}

func (r *Room) connectRequest(conn Conn) error {
	c := ConnectRequest{}

	err := conn.Read(&c)
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

func (r *Room) ping(conn Conn) error {
	defer conn.Close()

	pi := Ping{}

	err := conn.Read(&pi)
	if err != nil {
		return err
	}

	po := Pong{
		ReceivedAt: pi.SentAt,
	}

	return conn.Send(PongCmd, &po)
}
