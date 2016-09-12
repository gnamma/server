package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"time"
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

	if com.Command == ConnectRequestCmd {
		err = n.connection(buf, conn)
		if err != nil {
			log.Println("Unable to sustain connection command:", err)
			return
		}

		return
	} else if com.Command == PingCmd {
		err = n.ping(buf, conn)
		if err != nil {
			log.Println("Unable to reply to ping:", err)
			return
		}

		return
	}

	log.Println("Unknown command:", com.Command)
}

func (n *Networker) connection(buf *bytes.Buffer, conn net.Conn) error {
	c := ConnectRequest{}

	err := json.NewDecoder(buf).Decode(&c)
	if err != nil {
		return err
	}

	p := Player{
		Username: c.Username,
	}

	cv := ConnectVerdict{}

	if n.s.Room.CanJoin(p) {
		cv = ConnectVerdict{
			CanProceed: true,
			Message:    "Welcome to the server!",
		}
	} else {
		cv = ConnectVerdict{
			CanProceed: false,
			Message:    "Sorry. Connection rejected.",
		}
	}

	cv.Communication = Communication{
		Command: ConnectVerdictCmd,
		SentAt:  time.Now().UnixNano(),
	}

	out, err := json.Marshal(cv)
	if err != nil {
		return err
	}

	conn.Write(out)

	return nil
}

func (n *Networker) ping(buf *bytes.Buffer, conn net.Conn) error {
	pi := Ping{}

	err := json.NewDecoder(buf).Decode(&pi)
	if err != nil {
		return err
	}

	po := Pong{
		Communication: Communication{
			Command: PongCmd,
			SentAt:  time.Now().UnixNano(),
		},
		ReceivedAt: pi.SentAt,
	}

	out, err := json.Marshal(po)
	if err != nil {
		return err
	}

	conn.Write(out)

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

	return com, buf, nil
}
