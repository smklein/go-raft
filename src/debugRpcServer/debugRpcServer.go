package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

type Check int

func (t *Check) CheckRPC(in string, out *string) error {
	s := []string{in, "Delay"}
	fmt.Println("[Debug server] Sees input: ", in)
	*out = strings.Join(s, " ")
	return nil
}

func main() {
	check := new(Check)
	rpc.Register(check)
	rpc.HandleHTTP()
	fmt.Println("About to listen...")
	l, e := net.Listen("tcp", ":8081")
	if e != nil {
		fmt.Println("Error: ", e)
	} else {
		fmt.Println("No errors listening. About to serve...")
		http.Serve(l, nil)
	}
}
