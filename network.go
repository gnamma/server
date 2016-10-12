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
	"time"
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
		cc, err := c.Read()
		if err != nil {
			c.Raw.log.Println("Couldn't create child, closing connection:", err)
			c.Raw.Close()
			return
		}

		go func(cc *ChildConn) {
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

		}(cc)

		time.Sleep(time.Second / time.Duration(n.s.Opts.ReadSpeed))
	}
}

type ChildConn struct {
	buf *bytes.Buffer
	p   *ComConn
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
}

func NewComConn(c *Conn) *ComConn {
	cc := &ComConn{
		Raw: c,

		delayers: make([]chan struct{}, 0),
	}

	return cc
}

func (c *ComConn) Read() (*ChildConn, error) {
	buf, err := c.Raw.ReadRaw()
	if err != nil {
		return nil, err
	}

	return NewChildConn(buf, c), nil
}

// NOTE: Do note use in a configuration!
func (c *ComConn) ExpectAndRead(cmd string, v Preparer) error {
	cc, err := c.Read()
	if err != nil {
		return err
	}

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
	ch := c.wait()

	c.sendLock.Lock()
	err := c.Raw.Send(cmd, v)
	c.sendLock.Unlock()

	ch <- struct{}{}

	return err
}

func (c *ComConn) Done() {
	c.delayersLock.Lock()

	for _, ch := range c.delayers {
		ch <- struct{}{}
		<-ch
		close(ch)
	}

	c.delayers = c.delayers[0:0]

	c.delayersLock.Unlock()
}

func (c *ComConn) wait() chan struct{} {
	c.delayersLock.Lock()
	ch := make(chan struct{})

	c.delayers = append(c.delayers, ch)

	c.delayersLock.Unlock()

	<-ch

	return ch
}

func (c *ComConn) log() *log.Logger {
	return c.Raw.log
}

type Conn struct {
	NConn net.Conn
	ID    uint

	connBuf  *bufio.Reader
	connLock sync.RWMutex
	log      *log.Logger
}

func (c *Conn) ReadRaw() (*bytes.Buffer, error) {
	if c.connBuf == nil {
		c.connBuf = bufio.NewReader(c.NConn)
	}

	c.connLock.Lock()
	defer c.connLock.Unlock()

	lenSli, err := c.connBuf.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	lenSli = bytes.TrimSpace(lenSli)
	l, err := strconv.Atoi(string(lenSli))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, l)
	_, err = c.connBuf.Read(buf)
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

	return c.SendRaw(bytes.NewBuffer(out))
}

func (c *Conn) SendRaw(r io.Reader) error {
	rBuf := &bytes.Buffer{}
	_, err := io.Copy(rBuf, r)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	_, err = buf.WriteString(fmt.Sprintf("%v\n", rBuf.Len()))
	if err != nil {
		return err
	}

	_, err = rBuf.WriteTo(buf)
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
