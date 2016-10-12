package server

import (
	"log"
	"net"
	"net/http"
	"os"
)

type AssetServer struct {
	Dir   http.Dir
	Addr  string
	Ready chan struct{}

	l *log.Logger
}

func NewAssetServer(addr, dir string) *AssetServer {
	return &AssetServer{
		Addr:  addr,
		Dir:   http.Dir(dir),
		l:     log.New(os.Stdout, "assets: ", logFlags),
		Ready: make(chan struct{}),
	}
}

func (as *AssetServer) Listen() error {
	ln, err := net.Listen(ConnectionType, as.Addr)
	if err != nil {
		return err
	}

	go func() { as.Ready <- struct{}{} }()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err // Probably shouldn't break the server here...
		}

		c := &Conn{NConn: conn, log: as.l}
		err = as.Handle(c)
		if err != nil {
			return err
		}
	}
}

func (as *AssetServer) Handle(conn *Conn) error {
	keyBuf, err := conn.ReadRaw()
	if err != nil {
		return err
	}

	key := keyBuf.String()

	f, err := as.Dir.Open(key)
	if err != nil {
		return err
	}

	return conn.SendRaw(f)
}
