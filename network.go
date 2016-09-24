package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	ConnectionType = "tcp"
)

type Networker struct {
	s *Server
}

func (n *Networker) Handle(conn net.Conn) {
	c := Conn{nc: conn, l: n.s.log}

	for {
		com, err := c.ReadCom()
		if err != nil {
			c.l.Println("Couldn't read command:", err)
			c.Close()
			c.l.Println("Closing connection...")
			return
		}

		err = n.s.Room.Handle(com.Command, c)
		if err != nil {
			c.l.Println("Unable to respond:", err)
			continue
		}

		c.FlushCache()
	}
}

type Conn struct {
	nc  net.Conn
	buf *bytes.Buffer
	l   *log.Logger
}

func (c *Conn) ReadCom() (Communication, error) {
	com := Communication{}

	_, err := c.ReadRaw()
	if err != nil {
		return com, err
	}

	oldBuf := bytes.NewBuffer(c.buf.Bytes())
	err = c.Read(&com)
	if err == nil {
		c.buf = oldBuf
	}

	return com, err
}

func (c *Conn) Read(v Preparer) error {
	r, err := c.ReadRaw()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(r)
	if err != nil {
		return err
	}

	if buf.String() == "" {
		c.FlushCache()
		return ErrEmptyBuffer
	}

	err = json.Unmarshal(buf.Bytes(), v)
	if err == nil || err == io.EOF {
		c.FlushCache()
		return nil
	}

	return err
}

func (c *Conn) ReadRaw() (io.Reader, error) {
	if c.buf == nil {
		connBuf := bufio.NewReader(c.nc)

		lenStr, err := connBuf.ReadString('\n')
		if err != nil {
			return nil, err
		}

		lenStr = strings.TrimSpace(lenStr)

		l, err := strconv.Atoi(lenStr)
		if err != nil {
			return nil, err
		}

		buf := make([]byte, l)
		_, err = connBuf.Read(buf)
		if err != nil {
			return nil, err
		}

		c.buf = bytes.NewBuffer(buf)
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

	return c.SendRaw(bytes.NewBuffer(out))
}

func (c *Conn) SendRaw(r io.Reader) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}

	_, err = c.nc.Write([]byte(fmt.Sprintf("%v\n", buf.Len())))
	if err != nil {
		return err
	}

	_, err = io.Copy(c.nc, buf)
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
