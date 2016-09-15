package server

type Player struct {
	ID       uint
	Username string
}

func (p *Player) Valid() bool {
	return p.Username != ""
}
