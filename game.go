package server

type Game struct {
	s *Server
}

func (g *Game) CanJoin(p Player) bool {
	return p.Valid()
}
