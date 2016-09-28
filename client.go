package server

import (
	"io"
	"log"
	"net"
	"os"
)

// Mainly for testing. Perhaps bots too? I don't know.
type Client struct {
	Address  string
	Username string

	conn   *Conn
	player *Player
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.Address)
	if err != nil {
		return err
	}

	c.conn = &Conn{nc: conn, l: log.New(os.Stdout, "client: ", logFlags)}

	cr := ConnectRequest{
		Username: c.Username,
	}
	err = c.conn.Send(ConnectRequestCmd, &cr)
	if err != nil {
		return err
	}

	cv := ConnectVerdict{}
	err = c.conn.ExpectAndRead(ConnectVerdictCmd, &cv)
	if err != nil {
		return err
	}

	if !cv.CanProceed {
		return ErrClientRejected
	}

	c.player = &Player{
		ID:       cv.PlayerID,
		Username: c.Username,
		Nodes:    make(map[uint]*Node),
	}

	return nil
}

func (c *Client) Ping() (Pong, error) {
	po := Pong{}

	if c.conn == nil {
		return po, ErrClientNotConnected
	}

	err := c.conn.Send(PingCmd, &Ping{})
	if err != nil {
		return po, err
	}

	err = c.conn.ExpectAndRead(PongCmd, &po)
	return po, err
}

func (c *Client) Environment() (EnvironmentPackage, error) {
	ep := EnvironmentPackage{}

	if c.conn == nil {
		return ep, ErrClientNotConnected
	}

	err := c.conn.Send(EnvironmentRequestCmd, &EnvironmentRequest{})
	if err != nil {
		return ep, err
	}

	err = c.conn.ExpectAndRead(EnvironmentPackageCmd, &ep)
	return ep, err
}

func (c *Client) Asset(key string) (io.Reader, error) {
	if c.conn == nil {
		return nil, ErrClientNotConnected
	}

	err := c.conn.Send(AssetRequestCmd, &AssetRequest{Key: key})
	if err != nil {
		return nil, err
	}

	buf, err := c.conn.ReadRaw()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (c *Client) RegisterNode(n Node) error {
	if c.conn == nil {
		return ErrClientNotConnected
	}

	err := c.conn.Send(RegisterNodeCmd, &RegisterNode{
		Node: n,
		PID:  c.player.ID,
	})
	if err != nil {
		return err
	}

	rn := RegisteredNode{}
	err = c.conn.ExpectAndRead(RegisteredNodeCmd, &rn)
	if err != nil {
		return err
	}

	c.player.Nodes[rn.NID] = &n
	c.player.nodeCount = rn.NID

	return nil
}

func (c *Client) UpdateNode(n Node) error {
	if c.conn == nil {
		return ErrClientNotConnected
	}

	return c.conn.Send(UpdateNodeCmd, &UpdateNode{
		PID:      c.player.ID,
		NID:      n.ID,
		Position: n.Position,
		Rotation: n.Rotation,
	})
}
