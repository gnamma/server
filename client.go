package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
)

// Mainly for testing. Perhaps bots too? I don't know.
type Client struct {
	Addr       string
	AssetsAddr string
	Username   string

	player *Player
	conn   *Session
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return err
	}

	c.conn = &Session{
		Conn: &Conn{
			nc: conn,
			l:  log.New(os.Stdout, "client: ", logFlags),
		},
		chans: make(map[string][]chan *bytes.Buffer),
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

	conn := Conn{nc: nc}
	defer conn.Close()

	err = conn.SendRawString(key)
	if err != nil {
		return nil, err
	}

	return conn.ReadRaw()
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

type Session struct {
	*Conn

	chans map[string][]chan *bytes.Buffer
}

func (s *Session) ReadWait(cmd string, v Preparer) error {
	c := make(chan *bytes.Buffer)

	s.chans[cmd] = append(s.chans[cmd], c)

	buf := <-c

	conn := Conn{
		buf: buf,
	}

	return conn.Read(v)
}

func (s *Session) ReadLoop() error {
	for {
		com, err := s.Conn.ReadCom()
		if err != nil { // TODO: Account for raw coms, like file transfers
			return err
		}

		for _, c := range s.chans[com.Command] {
			c <- bytes.NewBuffer(s.Conn.buf.Bytes())
		}
	}

	return nil
}
