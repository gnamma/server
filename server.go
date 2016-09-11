package server

import "net"

type Options struct {
	Name        string
	Description string
	Address     string
}

type Server struct {
	Opts Options

	Netw *Networker
	Game *Game
}

func New(o Options) *Server {
	s := &Server{
		Opts: o,
	}

	s.Netw = &Networker{
		s: s,
	}

	s.Game = &Game{
		s: s,
	}

	return s
}

func (s *Server) Listen() error {
	ln, err := net.Listen(ConnectionType, s.Opts.Address)
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go s.Netw.Handle(conn)
	}
}
