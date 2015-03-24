package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

type Check int

type Server struct {
	Name string
	Address string
	Port int
}

type Rule map[string]string

type Config struct {
	Servers []Server
	Rules []Rule
}

func (t *Check) CheckRPC(in string, out *string) error {
	s := []string{in, "Delay"}
	fmt.Println("[Debug server] Sees input: ", in)
	*out = strings.Join(s, " ")
	return nil
}

var config Config

func main() {
	// Load config
	file, e := ioutil.ReadFile("config.yaml")
	if e != nil {
		fmt.Println("Error: ", e)
	}
	e = yaml.Unmarshal(file, &config)
	if e != nil {
		fmt.Println("Error: ", e)
	}
	fmt.Println("Config: ", config)

	// Start server
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
