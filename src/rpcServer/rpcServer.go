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
	*out = strings.Join(s, " ")
	return nil
}

func main() {
    check := new(Check)
	rpc.Register(check)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8080")
	if e == nil {
		go http.Serve(l, nil)
	}
	for {}
}