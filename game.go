package server

type Room struct {
	s *Server
}

func (g *Room) CanJoin(p Player) bool {
	return p.Valid()
}
