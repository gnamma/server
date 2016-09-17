package server

import (
	"bufio"
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
	c := Conn{nc: conn}

	com, err := c.ReadCom()
	if err != nil {
		log.Println("Couldn't read command:", err)
		return
	}

	err = n.s.Room.Respond(com.Command, c)
	if err != nil {
		log.Println("Unable to respond:", err)
		return
	}
}

type Conn struct {
	nc net.Conn

	buf *bytes.Buffer
}

func (c *Conn) ReadCom() (Communication, error) {
	buf := &bytes.Buffer{}
	cmd := &bytes.Buffer{}
	com := Communication{}

	r := bufio.NewReader(c.nc)

	msg, err := r.ReadString('\n')
	if err != nil {
		log.Println("coudlnt read from", err)
		return com, err
	}

	buf = bytes.NewBufferString(msg)
	c.buf = buf

	cmd = bytes.NewBuffer(buf.Bytes())

	err = json.NewDecoder(cmd).Decode(&com)
	if err != nil {
		return com, err
	}

	return com, nil
}

func (c *Conn) Read(v Preparer) error {
	if c.buf == nil {
		_, err := c.buf.ReadFrom(c.nc)
		if err != nil {
			return err
		}
	}

	err := json.NewDecoder(c.buf).Decode(v)
	c.buf = nil

	return err
}

func (c *Conn) Send(cmd string, v Preparer) error {
	v.Prepare(cmd)

	out, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = c.nc.Write(out)
	if err != nil {
		return err
	}

	_, err = c.nc.Write([]byte("\n"))
	return err
}

func (c *Conn) Close() error {
	return c.nc.Close()
}
