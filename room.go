package server

import (
	"log"
	"time"
)

type Room struct {
	*Dispatch

	s *Server

	players     map[uint]*Player
	playerCount uint
}

func NewRoom(s *Server) *Room {
	r := &Room{
		s:       s,
		players: make(map[uint]*Player),
	}

	r.Dispatch = &Dispatch{
		map[string]CommunicationHandler{
			PingCmd:               r.ping,
			ConnectRequestCmd:     r.connectRequest,
			EnvironmentRequestCmd: r.environmentRequest,
			RegisterNodeCmd:       r.registerNode,
			UpdateNodeCmd:         r.updateNode,
		},
	}

	return r
}

func (r *Room) StartUpdateLoop(fps float64) {
	wait := time.Second / time.Duration(fps)
	log.Println("Updating on an interval of:", wait)

	for {
		r.s.log.Println("Updating server...")
		for _, p := range r.players {
			p.Conn.Done()
		}

		time.Sleep(wait)
	}
}

func (r *Room) CanJoin(p *Player) bool {
	_, ok := r.players[p.ID]

	return p.Valid() && !ok
}

func (r *Room) Join(u string, c *ChildConn) (*Player, error) {
	p := &Player{
		Username: u,
		ID:       r.playerCount + 1, // Don't increment straight away so that to prevent an overflow.
		Nodes:    make(map[uint]*Node),
		Conn:     c.Parent(),
	}

	if !r.CanJoin(p) {
		return nil, ErrPlayerCantJoin
	}

	r.players[p.ID] = p

	r.playerCount += 1

	return p, nil
}

func (r *Room) ping(conn *ChildConn) error {
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

func (r *Room) connectRequest(conn *ChildConn) error {
	c := ConnectRequest{}

	err := conn.Read(&c)
	if err != nil {
		return err
	}

	cv := ConnectVerdict{
		CanProceed: true,
		Message:    "Welcome to the server!",
	}

	p, err := r.Join(c.Username, conn)
	if err != nil {
		cv = ConnectVerdict{
			CanProceed: false,
			Message:    "Sorry. Connection rejected.",
		}
	} else {
		cv.PlayerID = p.ID
		conn.log().Println("Connected player:", p)
	}

	return conn.Send(ConnectVerdictCmd, &cv)
}

func (r *Room) environmentRequest(conn *ChildConn) error {
	er := EnvironmentRequest{}

	err := conn.Read(&er)
	if err != nil {
		return err
	}

	ep := EnvironmentPackage{
		AssetKeys: map[string]string{
			"world": "room.gsml", // TODO: Do something here
		},
		Main: "world",
	}

	return conn.Send(EnvironmentPackageCmd, &ep)
}

func (r *Room) registerNode(conn *ChildConn) error {
	rn := RegisterNode{}

	err := conn.Read(&rn)
	if err != nil {
		return err
	}

	p, ok := r.players[rn.PID]
	if !ok {
		return ErrPlayerDoesntExist
	}

	nid, err := p.RegisterNode(rn.Node)
	if err != nil {
		return err
	}

	return conn.Send(RegisteredNodeCmd, &RegisteredNode{
		NID: nid,
	})
}

func (r *Room) updateNode(conn *ChildConn) error {
	un := UpdateNode{}

	err := conn.Read(&un)
	if err != nil {
		return err
	}

	p, ok := r.players[un.PID]
	if !ok {
		return ErrPlayerDoesntExist
	}

	n, ok := p.Nodes[un.NID]
	if !ok {
		return ErrNodeDoesntExist
	}

	n.Position = un.Position
	n.Rotation = un.Rotation

	return nil
}
