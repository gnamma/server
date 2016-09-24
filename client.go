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
