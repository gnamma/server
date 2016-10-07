package server

import (
	"log"
	"net"
	"os"
)

var (
	logFlags int = log.Lshortfile
)

type Options struct {
	Name        string
	Description string
	Addr        string

	AssetsDir  string
	AssetsAddr string
}

type Server struct {
	Opts Options

	Netw   *Networker
	Room   *Room
	Assets *AssetServer

	Ready chan struct{}

	log *log.Logger
}

func New(o Options) *Server {
	s := &Server{
		Opts:   o,
		Ready:  make(chan struct{}),
		log:    log.New(os.Stdout, "server: ", logFlags),
		Assets: NewAssetServer(o.AssetsAddr, o.AssetsDir),
	}

	s.Netw = &Networker{s: s}
	s.Room = NewRoom(s)

	return s
}

func (s *Server) Listen() error {
	ln, err := net.Listen(ConnectionType, s.Opts.Addr)
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

func (s *Server) Go() error {
	go s.Assets.Listen()

	return s.Listen()
}
