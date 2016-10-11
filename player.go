package server

import "sync"

type NodeType uint

const (
	HeadNode NodeType = iota + 1
	ArmNode
)

type Player struct {
	ID       uint
	Username string

	Conn *ComConn

	Nodes     map[uint]*Node
	nodesLock sync.RWMutex
	nodeCount uint
}

func (p *Player) Valid() bool {
	return p.Username != ""
}

func (p *Player) RegisterNode(n Node) (uint, error) {
	id := p.nodeCount + 1

	p.nodesLock.Lock()
	_, ok := p.Nodes[id]
	if ok {
		return 0, ErrNodeAlreadyExists
	}

	p.Nodes[id] = &n
	p.nodeCount += 1

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
	X float64
	Y float64
	Z float64
}
