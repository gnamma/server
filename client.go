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

	chans map[string]chan *ChildConn
}

func (c *Client) readLoop() {
	c.chans = make(map[string]chan *ChildConn)

	for {
		cc, err := c.conn.Read()
		if err != nil {
			c.conn.log().Println("Error in client read loop:", err)
			continue
		}

		go func(cc *ChildConn) {
			com, err := cc.Com()
			if err != nil {
				c.conn.log().Println("Error reading com:", err)
				return
			}

			ch := c.populateChan(com.Command)

			ch <- cc
		}(cc)
	}
}

func (c *Client) WaitFor(cmd string) *ChildConn {
	ch := c.populateChan(cmd)

	return <-ch
}

func (c *Client) populateChan(cmd string) chan *ChildConn {
	ch, ok := c.chans[cmd]

	if !ok {
		c.chans[cmd] = make(chan *ChildConn)
		ch = c.chans[cmd]
	}

	return ch
}

func (c *Client) UpdateLoop() {
	go c.readLoop()

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
	err = c.ExpectAndRead(ConnectVerdictCmd, &cv)
	if err != nil {
		return err
	}

	if !cv.CanProceed {
		return ErrClientRejected
	}

	c.player = &Player{
		ID:       cv.PlayerID,
		Username: c.Username,
		nodesMap: make(map[uint]*Node),
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

	err = c.ExpectAndRead(PongCmd, &po)
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

	err = c.ExpectAndRead(EnvironmentPackageCmd, &ep)
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
	err = c.ExpectAndRead(RegisteredNodeCmd, &rn)
	if err != nil {
		return err
	}

	n.ID = rn.NID
	n.PID = c.player.ID

	c.player.nodesMap[rn.NID] = n
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

func (c *Client) ExpectAndRead(cmd string, v Preparer) error {
	cc := c.WaitFor(cmd)

	return cc.Read(v)
}
