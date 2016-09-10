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

	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(conn)
	if err != nil {
		log.Println("Unable to read properly:", err)
		return
	}

	cmd := bytes.NewBuffer(buf.Bytes())
	if err != nil {
		log.Println("Unable to copy:", err)
		return
	}

	com := Communication{}

	err = json.NewDecoder(cmd).Decode(&com)
	if err != nil {
		log.Println("Unable to decode JSON:", err)
		return
	}

	if com.Command != "connection" {
		fmt.Println("Unknown command:", com.Command)
		return
	}

	err = n.connection(buf, conn)
	if err != nil {
		log.Println("Unable to sustain connection command:", err)
		return
	}
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
