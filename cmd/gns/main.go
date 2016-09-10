package main

import (
	"github.com/gnamma/server"
	"log"
)

func main() {
	s := server.New(server.Options{
		Name:        "Bar",
		Description: "Yeah. This seems appropriate?",
		Address:     ":3000",
	})

	log.Println("Starting Gnamma server...")

	log.Printf("Name: %v, Description: %v\n", s.Opts.Name, s.Opts.Description)

	err := s.Listen()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Exiting")
}
