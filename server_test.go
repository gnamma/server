package server

import (
	"os"
	"testing"
)

var (
	address = "localhost:3445"

	server *Server
	client *Client
)

func TestMain(m *testing.M) {
	server = New(Options{
		Name:        "Test Server",
		Description: "Used for testing",
		Address:     address,
	})

	client = &Client{
		Address:  address,
		Username: "parzival", // Ten points to Ravenclaw if someone gets this reference.
	}

	go server.Listen()

	<-server.Ready
	os.Exit(m.Run())
}

func TestConnect(t *testing.T) {
	err := client.Connect()
	if err != nil {
		t.Fatalf("Client could not connect to the server", err)
	}
}

func TestPing(t *testing.T) {
	_, err := client.Ping()
	if err != nil {
		t.Fatalf("Client could not ping the server", err)
	}
}

func TestRequestEnvironment(t *testing.T) {
	er, err := client.Environment()
	if err != nil {
		t.Fatalf("Client could not get environment from server", err)
	}

	if er.Main != "world" {
		t.Fatalf("Wrong main file, expected 'world', got '%v'", er.Main)
	}
}
