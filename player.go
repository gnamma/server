package server

type Player struct {
	Username string
}

func (p *Player) Valid() bool {
	return p.Username != ""
}
