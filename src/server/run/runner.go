package main

import (
	"fmt"
	"os"
	"server"
)

func main() {
	fmt.Println("[MAIN] Server Started: ", os.Args[1])

	s := server.CreateRaftServer(os.Args[1])

	if s == nil {
		fmt.Println("[MAIN] Could not create server")
	}
}
