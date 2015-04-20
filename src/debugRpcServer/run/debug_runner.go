package main

import (
	"config"
	"debugRpcServer"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

func main() {
	fmt.Println("[MAIN] Debug Server Started...")

	if !config.LoadConfig(&debugRpcServer.Cf) {
		fmt.Println("[MAIN] Couldn't load config...")
		return
	} else {
		fmt.Println("[MAIN] Config: ", debugRpcServer.Cf)
	}

	// Start server
	check := new(debugRpcServer.Check)
	rpc.Register(check)
	rpc.HandleHTTP()
	fmt.Println("About to listen...")
	var port int
	for _, s := range debugRpcServer.Cf.Servers {
		if s.Name == "debugServer" {
			port = s.Port
		}
	}
	l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		fmt.Println("Error: ", e)
	} else {
		fmt.Println("No errors listening. About to serve...")
		http.Serve(l, nil)
	}

}
