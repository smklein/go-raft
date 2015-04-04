package server

import (
	"fmt"
	"raftRpc"
)

// TODO Get rid of these...
var (
	serverAddr1 = "localhost:8081"
	serverAddr2 = "localhost:9002"
	serverAddr3 = "localhost:9003"
	serverAddr4 = "localhost:9004"
)

func main() {
	fmt.Println("Hi from main!")
	rc, err := raftRpc.DialHTTP("tcp", serverAddr1)
	if err != nil {
		fmt.Println("Cannot dial: " + serverAddr1)
		return
	}
	fmt.Println("Successfully dialed: ", serverAddr1)
	args := "this is the args"
	var reply string
	err = rc.Call("Check.CheckRPC", args, &reply)
	if err != nil {
		fmt.Println("Whoops! We can't dial checkrpc.")
		return
	}
	fmt.Println("Main sees reply from server: " + reply)
	rc.Close()
}

func GimmeTrue() bool {
	return true
}
