package server

import "log"

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

func (r *Room) CanJoin(p *Player) bool {
	_, ok := r.players[p.ID]

	return p.Valid() && !ok
}

func (r *Room) Join(u string) (*Player, error) {
	p := &Player{
		Username: u,
		ID:       r.playerCount + 1, // Don't increment straight away so that to prevent an overflow.
		Nodes:    make(map[uint]*Node),
	}

	if !r.CanJoin(p) {
		return nil, ErrPlayerCantJoin
	}

	r.players[p.ID] = p

	r.playerCount += 1

	return p, nil
}

func (r *Room) ping(conn Conn) error {
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
	} else {
		cv.PlayerID = p.ID
	}

	log.Println("Connected player:", p)

	return conn.Send(ConnectVerdictCmd, &cv)
}

func (r *Room) environmentRequest(conn Conn) error {
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

func (r *Room) registerNode(conn Conn) error {
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

func (r *Room) updateNode(conn Conn) error {
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
