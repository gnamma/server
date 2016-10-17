package main

import (
	"flag"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gnamma/server"
)

var (
	addr       = flag.String("address", "localhost:3000", "The address for the server you want to connect to")
	assetsAddr = flag.String("assets-address", "localhost:3001", "The address for the specific address server you want to listen on")
	username   = flag.String("username", "reverb", "The username this bot will take")

	client *server.Client
)

func main() {
	flag.Parse()

	client = &server.Client{
		Addr:       *addr,
		AssetsAddr: *assetsAddr,
		Username:   *username,
	}

	err := client.Connect()
	if err != nil {
		log.Fatal("Couldn't connect to server:", err)
		return
	}

	log.Println(*client)

	var nodes = []*server.Node{
		{
			Type:     server.HeadNode,
			Label:    "Your head, bro!",
			Asset:    "box",
			Position: server.Point{0, 2, 0},
		},
		{
			Type:     server.ArmNode,
			Label:    "This is your arm, sis!",
			Asset:    "box",
			Position: server.Point{-1, 1, 0},
		},
		{
			Type:     server.ArmNode,
			Label:    "This is your arm, you!",
			Asset:    "box",
			Position: server.Point{1, 1, 0},
		},
	}

	var wg sync.WaitGroup

	for _, n := range nodes {
		wg.Add(1)

		go func(n *server.Node) {
			defer wg.Done()

			err := client.RegisterNode(n)
			if err != nil {
				log.Println("Unable to register node: ", err)
				return
			}
		}(n)
	}

	wg.Wait()

	move(nodes)
}

func move(nodes []*server.Node) {
	speed := float64(math.Pi/180) * 5 // Want to move 1 radian each iteration
	x := float64(0)

	wait := time.Second / time.Duration(10)

	for {
		var wg sync.WaitGroup

		for _, n := range nodes {
			wg.Add(1)

			go func(n *server.Node) {
				defer wg.Done()
				n.Position.Z = math.Sin(x)

				log.Println("at:", n.Position.Z)

				err := client.UpdateNode(*n)
				if err != nil {
					log.Fatal("Couldn't update node:", err)
					return
				}
			}(n)
		}

		x += speed
		wg.Wait()

		time.Sleep(wait)
	}

}
