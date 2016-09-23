package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
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

	for {
		com, err := c.ReadCom()
		if err != nil {
			log.Println("Couldn't read command:", err)
			return
		}

		err = n.s.Room.Handle(com.Command, c)
		if err != nil {
			log.Println("Unable to respond:", err)
			return
		}
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
	buf, err := c.ReadRaw()
	if err != nil {
		return err
	}

	err = json.NewDecoder(buf).Decode(v)

	c.FlushCache()

	return err
}

func (c *Conn) ReadRaw() (io.Reader, error) {
	if c.buf == nil {
		buf := bufio.NewReader(c.nc)
		s, err := buf.ReadString('\n')
		if err != nil {
			return nil, err
		}

		c.buf = bytes.NewBufferString(s)
	}

	return c.buf, nil
}

func (c *Conn) FlushCache() {
	c.buf = nil
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

func (c *Conn) SendRaw(r io.Reader) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}

	buf.WriteRune('\n')

	io.Copy(c.nc, buf)

	return err
}

func (c *Conn) Close() error {
	return c.nc.Close()
}

func (c *Conn) Expect(cmd string) error {
	com, err := c.ReadCom()
	if err != nil {
		return err
	}

	if com.Command != cmd {
		return ErrUnexpectedCom
	}

	return nil
}

func (c *Conn) ExpectAndRead(cmd string, v Preparer) error {
	err := c.Expect(cmd)
	if err != nil {
		return err
	}

	return c.Read(v)
}
