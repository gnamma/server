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
	FPS         float64

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
	if o.FPS == 0 {
		o.FPS = 60
	}

	s := &Server{
		Opts:   o,
		Ready:  make(chan struct{}),
		Assets: NewAssetServer(o.AssetsAddr, o.AssetsDir),

		log: log.New(os.Stdout, "server: ", logFlags),
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

	go s.Room.StartUpdateLoop(s.Opts.FPS)

	ids := uint(0)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go s.Netw.Handle(conn, ids)

		ids += 1
	}
}

func (s *Server) Go() error {
	go s.Assets.Listen()

	return s.Listen()
}
