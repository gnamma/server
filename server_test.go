package server

import (
	"bytes"
	"log"
	"os"
	"sync"
	"testing"
)

var (
	serverAddr = "localhost:3445"
	assetsAddr = "localhost:3554"
	files      = "test"

	server *Server
	client *Client
)

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	server = New(Options{
		Name:        "Test Server",
		Description: "Used for testing",
		Addr:        serverAddr,
		AssetsDir:   files,
		AssetsAddr:  assetsAddr,
	})

	client = &Client{
		Addr:       serverAddr,
		Username:   "parzival", // Ten points to Ravenclaw if someone gets this reference.
		AssetsAddr: assetsAddr,
	}

	go server.Go()

	<-server.Ready
	<-server.Assets.Ready

	os.Exit(m.Run())
}

func TestConnect(t *testing.T) {
	err := client.Connect()
	if err != nil {
		t.Fatalf("Client could not connect to the server", err)
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

func TestAssetRequest(t *testing.T) {
	r, err := client.Asset("main")
	if err != nil {
		t.Fatal("Client could not retrieve asset from server", err)
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatal("Couldn't read from asset buffer")
	}

	if buf.String() != "<room></room>\n" {
		t.Fatalf("Asset is not the same!")
	}
}

func TestPing(t *testing.T) {
	_, err := client.Ping()
	if err != nil {
		t.Fatal("Client could not ping the server:", err)
	}
}

func TestPings(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			t.Logf("sending #%d", i)

			_, err := client.Ping()
			if err != nil {
				t.Fatal("Client could not ping server!")
			}

			t.Logf("got #%d", i)
		}(i)
	}

	wg.Wait()
}

func TestNodes(t *testing.T) {
	nodes := []Node{
		{
			Type:     ArmNode,
			Position: Point{1, 1, 0},
			Rotation: Point{},
			Label:    "your right arm!",
		},
		{
			Type:     ArmNode,
			Position: Point{-1, 1, 0},
			Rotation: Point{},
			Label:    "your left arm!",
		},
		{
			Type:     HeadNode,
			Position: Point{0, 2, 0},
			Rotation: Point{},
			Label:    "your head!",
		},
	}

	var wg sync.WaitGroup

	for _, n := range nodes {
		wg.Add(1)
		go func(n Node) {
			t.Log("sending:", n.Label)
			defer wg.Done()

			err := client.RegisterNode(n)
			if err != nil {
				t.Fatalf("Unable to register node '%s': %v", n.Label, err)
			}

			t.Log("got:", n.Label)
		}(n)
	}

	wg.Wait()
}
