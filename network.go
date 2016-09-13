package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
)

const (
	ConnectionType = "tcp"
)

type Networker struct {
	s *Server
}

func (n *Networker) Handle(conn net.Conn) {
	defer conn.Close()

	com, buf, err := n.peek(conn)
	if err != nil {
		log.Println("Could not peek command:", err)
		return
	}

	err = n.s.Room.Respond(com.Command, buf, Conn{conn})
	if err != nil {
		log.Println("Unable to respond:", err)
		return
	}
}

func (n *Networker) peek(conn net.Conn) (Communication, *bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	cmd := &bytes.Buffer{}
	com := Communication{}

	_, err := buf.ReadFrom(conn)
	if err != nil {
		return com, cmd, err
	}

	cmd = bytes.NewBuffer(buf.Bytes())

	err = json.NewDecoder(cmd).Decode(&com)
	if err != nil {
		return com, cmd, err
	}

	return com, buf, nil
}

type Conn struct {
	net.Conn
}

func (c *Conn) Send(cmd string, v Preparer) error {
	v.Prepare(cmd)

	out, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = c.Write(out)
	return err
}
