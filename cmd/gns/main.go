package main

import (
	"flag"
	"log"

	"github.com/gnamma/server"
)

var (
	name        = flag.String("name", "server", "The name of the server which you want to host")
	description = flag.String("description", "Greetings, traveller!", "A short description of the server")
	address     = flag.String("address", ":3000", "The address which you want to host the server on, etc localhost:3000")
	assets      = flag.String("assets", "cmd/gns/example", "The path of where the files for the server are kept")
	assetsAddr  = flag.String("assets-addr", ":3001", "The address of where to host the asset server")
)

func main() {
	flag.Parse()

	s := server.New(server.Options{
		Name:        *name,
		Description: *description,
		Addr:        *address,
		AssetsDir:   *assets,
		AssetsAddr:  *assetsAddr,
	})

	log.Println("Starting Gnamma server...")

	log.Printf("Name: %v, Description: %v\n", s.Opts.Name, s.Opts.Description)

	err := s.Go()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Exiting")
}
