package server

import (
	"log"
	"net"
	"os"
)

var (
	logFlags int = log.LstdFlags | log.Lshortfile
)

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

	Ready chan struct{}

	log *log.Logger
}

func New(o Options) *Server {
	s := &Server{
		Opts:  o,
		Ready: make(chan struct{}),
		log:   log.New(os.Stdout, "server: ", logFlags),
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

	go func() { s.Ready <- struct{}{} }()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go s.Netw.Handle(conn)
	}
}
