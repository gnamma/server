package server

import (
	"io"
	"log"
	"net"
	"os"
	"time"
)

// Mainly for testing. Perhaps bots too? I don't know.
type Client struct {
	Addr       string
	AssetsAddr string
	Username   string

	// ReadSpeed is the amount of times the client should read the server.
	ReadSpeed float64

	player *Player
	conn   *ComConn
}

func (c *Client) UpdateLoop() {
	if c.ReadSpeed == 0 {
		c.ReadSpeed = DefaultReadSpeed
	}

	wait := time.Second / time.Duration(c.ReadSpeed)

	for {
		c.conn.Done()
		time.Sleep(wait)
	}
}

func (c *Client) setup() error {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return err
	}

	c.conn = NewComConn(&Conn{
		NConn: conn,
		log:   log.New(os.Stdout, "client: ", logFlags),
	})

	go c.UpdateLoop()

	return nil
}

func (c *Client) Connect() error {
	err := c.setup()
	if err != nil {
		return err
	}

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
	nc, err := net.Dial("tcp", c.AssetsAddr)
	if err != nil {
		return nil, err
	}

	conn := Conn{NConn: nc, log: log.New(os.Stdout, "asset cli: ", logFlags)}
	defer conn.Close()

	err = conn.SendRawString(key)
	if err != nil {
		return nil, err
	}

	return conn.ReadRaw()
}

func (c *Client) RegisterNode(n *Node) error {
	if c.conn == nil {
		return ErrClientNotConnected
	}

	err := c.conn.Send(RegisterNodeCmd, &RegisterNode{
		Node: *n,
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

	n.ID = rn.NID
	n.PID = c.player.ID

	c.player.Nodes[rn.NID] = n
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
