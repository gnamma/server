package server

import "net"

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

	c.conn = &Conn{nc: conn}

	cr := ConnectRequest{
		Username: c.Username,
	}

	cv := ConnectVerdict{}

	err = c.conn.Send(ConnectRequestCmd, &cr)
	if err != nil {
		return err
	}

	err = c.conn.Expect(ConnectVerdictCmd)
	if err != nil {
		return err
	}

	err = c.conn.Read(&cv)
	if err != nil {
		return err
	}

	if !cv.CanProceed {
		return ErrClientRejected
	}

	return nil
}
