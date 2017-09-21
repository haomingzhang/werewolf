package main

import (
	"github.com/haomingzhang/werewolf/client"
	"github.com/haomingzhang/werewolf/game"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		runLocal()
	}

	switch args[0] {
	case game.ServerMode:
		runServer()
		return
	case game.ClientMode:
		runClient(args[1])
		return
	case game.LocalMode:
		fallthrough
	default:
		runLocal()
	}
}

func runLocal() {
	playBeginGame()
	gs := &game.GameServer{}
	gs.Controller = game.CreateController(game.LocalMode)
	gs.Start()
}

func runServer() {
	gs := &game.GameServer{}
	gs.Controller = game.CreateController(game.ServerMode)
	gs.Start()
}

func runClient(serverHost string) {
	playBeginGame()
	c, err := client.CreateWerewolfClient(serverHost)
	if err != nil {
		log.Fatal(err)
		return
	}
	c.Start()
}

func playBeginGame() {
	game.PlayAudio("serverBegin.mpg")
}
