package main

import (
	"fmt"
	"os"
	"server"
)

func main() {
	fmt.Fprintf(os.Stdout, "[MAIN] Server Started: %s\n", os.Args[1])

	s := server.CreateRaftServer(os.Args[1])

	if s == nil {
		fmt.Println("[MAIN] Could not create server")
	}
}
