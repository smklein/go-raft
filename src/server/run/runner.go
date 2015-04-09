package main

import (
	"config"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"server"
	"strconv"
)

func main() {
	fmt.Println("[MAIN] Server Started: ", os.Args[1])

	if !config.LoadConfig(&server.Cf) {
		fmt.Println("[MAIN] Couldn't load config...")
		return
	} else {
		fmt.Println("[MAIN] Config: ", server.Cf)
	}

	// Start server
	rs := new(server.RaftServer)
	rpc.Register(rs)
	rpc.HandleHTTP()
	fmt.Println("About to listen...")
	var port int
	for _, s := range server.Cf.Servers {
		if s.Name == os.Args[1] {
			port = s.Port
			fmt.Println("[MAIN] Server port found: ", port)
		}
	}
	if port == 0 {
		fmt.Println("[MAIN] Server name not identified by config file.")
		os.Exit(1)
	}

	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		fmt.Println("Error: ", e)
	} else {
		fmt.Println("No errors listening. About to serve...")
		http.Serve(l, nil)
	}
}
