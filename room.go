package server

import (
	"log"
	"time"
)

type Room struct {
	*Dispatch

	Broadcast chan Broadcast

	s *Server

	players     map[uint]*Player
	playerCount uint
}

func NewRoom(s *Server) *Room {
	r := &Room{
		s:         s,
		players:   make(map[uint]*Player),
		Broadcast: make(chan Broadcast),
	}

	r.Dispatch = &Dispatch{
		map[string]CommunicationHandler{
			PingCmd:               r.ping,
			ConnectRequestCmd:     r.connectRequest,
			EnvironmentRequestCmd: r.environmentRequest,
			RegisterNodeCmd:       r.registerNode,
			UpdateNodeCmd:         r.updateNode,
			RegisteredAllNodesCmd: r.registeredAllNodes,
		},
	}

	go r.broadcastLoop()

	return r
}

func (r *Room) StartUpdateLoop() {
	wait := time.Second / time.Duration(r.s.Opts.WriteSpeed)
	log.Println("Updating on an interval of:", wait)

	for {
		for k, p := range r.players {
			p.Conn.Done()

			if p.Conn.Closed {
				log.Println("Hey we're closing!")
				r.Broadcast <- Broadcast{
					Cmd: LeaveRoomCmd,
					Com: &LeaveRoom{PID: p.ID},
				}

				log.Println("CLosed!")

				delete(r.players, k)
			}
		}

		time.Sleep(wait)
	}
}

func (r *Room) broadcastLoop() {
	for {
		b := <-r.Broadcast

		if b.Cmd != UpdateNodeCmd {
			log.Println(b.Cmd)
		}

		for _, p := range r.players {
			go func(p *Player) { // This is not going to garbage collect well...
				if p.Conn.Closed {
					return
				}

				log.Println("Sending...")
				err := p.Conn.Send(b.Cmd, b.Com)
				log.Println("Sent!")
				if err != nil {
					p.Conn.Close()

					return
				}
			}(p)
		}

		log.Println("Got to the end of this broadcast!")
	}
}

func (r *Room) Player(pid uint) (*Player, error) {
	p, ok := r.players[pid]
	if !ok {
		return nil, ErrPlayerDoesntExist
	}

	return p, nil
}

func (r *Room) CanJoin(p *Player) bool {
	_, ok := r.players[p.ID]

	return p.Valid() && !ok
}

func (r *Room) Join(u string, c *ChildConn) (*Player, error) {
	p := &Player{
		Username: u,
		ID:       r.playerCount + 1, // Don't increment straight away so that to prevent an overflow.
		nodesMap: make(map[uint]*Node),
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

		var ps []Player

		for _, p := range r.players {
			if p.ID == cv.PlayerID {
				continue
			}

			ps = append(ps, *p)
		}

		cv.Players = ps
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

	p, err := r.Player(un.PID)
	if err != nil {
		return err
	}

	n, ok := p.nodesMap[un.NID]
	if !ok {
		return ErrNodeDoesntExist
	}

	n.Position = un.Position
	n.Rotation = un.Rotation

	r.Broadcast <- Broadcast{
		Cmd: UpdateNodeCmd,
		Com: &un,
	}

	return nil
}

func (r *Room) registeredAllNodes(conn *ChildConn) error {
	ran := RegisteredAllNodes{}

	err := conn.Read(&ran)
	if err != nil {
		return err
	}

	p, err := r.Player(ran.PID)
	if err != nil {
		return err
	}

	r.Broadcast <- Broadcast{
		Cmd: JoinRoomCmd,
		Com: &JoinRoom{
			Player: *p,
		},
		From: p.ID,
	}

	return nil
}

type Broadcast struct {
	Cmd  string
	Com  Preparer
	From uint
}
