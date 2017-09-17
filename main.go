package main

import (
	"github.com/haomingzhang/werewolf/game"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("/bin/bash", "-c", "afplay ./audio/serverBegin.mpg")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
	gs := &game.GameServer{}
	gs.Start()
}
