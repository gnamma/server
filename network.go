package server

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	if com.Command == "connection" {
		err = n.connection(buf, conn)
		if err != nil {
			log.Println("Unable to sustain connection command:", err)
			return
		}
	}

	log.Println("Unknown command:", com.Command)
	return
}

func (n *Networker) connection(buf *bytes.Buffer, conn net.Conn) error {
	c := Connection{}

	err := json.NewDecoder(buf).Decode(&c)
	if err != nil {
		return err
	}

	conn.Write([]byte(fmt.Sprintf("Hello, %v!", c.Username)))

	return nil
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

	return com, cmd, nil
}
