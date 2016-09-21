package server

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

type Assets struct {
	s *Server

	Dir string
}

func (a *Assets) Get(key string) (io.Reader, error) {
	f, err := os.Open(filepath.Join(a.Dir, key))
	if err != nil {
		return nil, err
	}

	return bufio.NewReader(f), nil
}
