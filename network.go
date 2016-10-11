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
	"sync"
)

const (
	ConnectionType = "tcp"
)

type Networker struct {
	s *Server
}

func (n *Networker) Handle(conn net.Conn, id uint) {
	rc := &Conn{
		NConn: conn,
		ID:    id,

		log: n.s.log,
	}

	c := NewComConn(rc)

	for {
		cc := c.Read()
		com, err := cc.Com()
		if err != nil {
			c.Raw.log.Println("Couldn't read command, closing connection:", err)
			c.Raw.Close()
			return
		}

		err = n.s.Room.Handle(com.Command, cc)
		if err != nil {
			c.Raw.log.Println("Couldn't handle com:", err)
		}
	}
}

type ChildConn struct {
	buf *bytes.Buffer

	p *ComConn
}

func NewChildConn(b *bytes.Buffer, p *ComConn) *ChildConn {
	return &ChildConn{
		buf: b,
		p:   p,
	}
}

func (cc *ChildConn) Com() (Communication, error) {
	com := Communication{}

	err := json.Unmarshal(cc.buf.Bytes(), &com)
	return com, err
}

func (cc *ChildConn) Read(p Preparer) error {
	if cc.buf.Len() == 0 {
		return ErrEmptyBuffer
	}

	err := json.Unmarshal(cc.buf.Bytes(), p)
	if err != nil {
		return err
	}

	cc.buf.Reset()
	return nil
}

func (cc *ChildConn) Send(cmd string, v Preparer) error {
	return cc.Parent().Send(cmd, v)
}

func (cc *ChildConn) Parent() *ComConn {
	return cc.p
}

func (cc *ChildConn) log() *log.Logger {
	return cc.Parent().log()
}

type ComConn struct {
	Raw *Conn

	delayers     []chan struct{}
	delayersLock sync.RWMutex
	sendLock     sync.RWMutex

	main chan *bytes.Buffer
}

func NewComConn(c *Conn) *ComConn {
	cc := &ComConn{
		Raw: c,

		delayers: make([]chan struct{}, 0),
		main:     make(chan *bytes.Buffer),
	}

	go cc.ReadLoop()

	return cc
}

func (c *ComConn) Read() *ChildConn {
	buf := <-c.main

	return NewChildConn(buf, c)
}

func (c *ComConn) ReadLoop() error {
	for {
		b, err := c.Raw.ReadRaw()
		if err != nil {
			c.log().Println("Couldn't read message:", err)
			continue
		}

		c.log().Println("read:", b.String())
		c.main <- b
		c.log().Println("delivered to handler")
	}
}

// NOTE: Do note use in a configuration!
func (c *ComConn) ExpectAndRead(cmd string, v Preparer) error {
	cc := c.Read()
	com, err := cc.Com()
	if err != nil {
		return err
	}

	if com.Command != cmd {
		return ErrUnexpectedCom
	}

	err = cc.Read(v)

	return err
}

func (c *ComConn) Send(cmd string, v Preparer) error {
	c.wait()

	c.sendLock.Lock()
	c.log().Println("started sending")
	err := c.Raw.Send(cmd, v)
	c.log().Println("finished sending")
	c.sendLock.Unlock()

	return err
}

func (c *ComConn) Done() {
	c.delayersLock.Lock()

	for _, ch := range c.delayers {
		ch <- struct{}{}
	}

	c.delayers = c.delayers[0:0] //make([]chan struct{}, 0)

	c.delayersLock.Unlock()
}

func (c *ComConn) wait() {
	c.delayersLock.Lock()
	ch := make(chan struct{})

	c.delayers = append(c.delayers, ch)

	c.delayersLock.Unlock()

	<-ch
}

func (c *ComConn) log() *log.Logger {
	return c.Raw.log
}

type Conn struct {
	NConn net.Conn
	ID    uint

	log *log.Logger
}

func (c *Conn) ReadRaw() (*bytes.Buffer, error) {
	connBuf := bufio.NewReader(c.NConn)

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

	return bytes.NewBuffer(buf), nil
}

func (c *Conn) Read(v Preparer) error {
	r, err := c.ReadRaw()
	if err != nil {
		return err
	}

	err = json.Unmarshal(r.Bytes(), v)
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) ReadCom() (Communication, error) {
	com := Communication{}

	err := c.Read(&com)
	return com, err
}

func (c *Conn) Send(cmd string, v Preparer) error {
	v.Prepare(cmd)

	out, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.log.Println("being sent:", string(out))

	return c.SendRaw(bytes.NewBuffer(out))
}

func (c *Conn) SendRaw(r io.Reader) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}

	_, err = c.NConn.Write([]byte(fmt.Sprintf("%v\n", buf.Len())))
	if err != nil {
		return err
	}

	_, err = io.Copy(c.NConn, buf)
	return err
}

func (c *Conn) SendRawString(s string) error {
	return c.SendRaw(strings.NewReader(s))
}

func (c *Conn) Close() error {
	return c.NConn.Close()
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
