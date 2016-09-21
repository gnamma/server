package server

import "net"

type Options struct {
	Name        string
	Description string
	Address     string

	AssetDir string
}

type Server struct {
	Opts Options

	Netw   *Networker
	Room   *Room
	Assets *Assets
}

func New(o Options) *Server {
	s := &Server{
		Opts: o,
	}

	s.Netw = &Networker{s: s}
	s.Room = NewRoom(s)
	s.Assets = &Assets{s: s, Dir: o.AssetDir}

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
