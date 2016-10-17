package server

import "sync"

type NodeType uint

const (
	HeadNode NodeType = iota + 1
	ArmNode
)

type Player struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`

	Conn *ComConn `json:"-"`

	// TODO: Neaten up this whole system
	Nodes     []*Node        `json:"nodes"`
	nodesMap  map[uint]*Node // Map for quick access
	nodesLock sync.RWMutex
	nodeCount uint
}

func (p *Player) Valid() bool {
	return p.Username != ""
}

func (p *Player) RegisterNode(n Node) (uint, error) {
	id := p.nodeCount + 1

	p.nodesLock.Lock()
	_, ok := p.nodesMap[id]
	if ok {
		return 0, ErrNodeAlreadyExists
	}

	p.Nodes = append(p.Nodes, &n)

	p.nodesMap[id] = &n
	p.nodeCount += 1

	n.PID = p.ID
	n.ID = id

	p.nodesLock.Unlock()

	return id, nil
}

type Node struct {
	ID       uint     `json:"id"`
	Type     NodeType `json:"type"`
	PID      uint     `json:"pid"`
	Position Point    `json:"position"`
	Rotation Point    `json:"rotation"`
	Asset    string   `json:"asset"`
	Label    string   `json:"label"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}
