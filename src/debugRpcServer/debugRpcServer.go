package debugRpcServer

import (
	"errors"
	"fmt"
	"strings"
)

type Check int

type ServerConnection struct {
	// Serves as key for rule map.
	// Identifies connection between input server and output server.
	Input  string
	Output string
}

type Behavior struct {
	Drop  bool
	Delay bool
}

type Config struct {
	Servers []string
}

var rules map[ServerConnection]*Behavior
var config Config

func (t *Check) CheckRPC(in string, out *string) error {
	s := []string{in, "Delay"}
	fmt.Println("[Debug server] Sees input: ", in)
	*out = strings.Join(s, " ")
	return nil
}

type RuleCommandRpcInput struct {
	input    string
	output   string
	behavior string
	on       bool
}

func serverExists(name string) bool {
	for _, server := range config.Servers {
		if server == name {
			return true
		}
	}
	return false
}

func addRuleToServer(in RuleCommandRpcInput, inServer, outServer string) {
	sc := ServerConnection{inServer, outServer}
	var behavior *Behavior
	if rules[sc] != nil {
		behavior = rules[sc]
	} else {
		behavior = &Behavior{} // Behaviors default to false
	}

	if in.behavior == "delay" {
		behavior.Delay = in.on
	}
	if in.behavior == "drop" {
		behavior.Drop = in.on
	}

	rules[sc] = behavior
}

func (t *Check) AddRule(in RuleCommandRpcInput, out *bool) error {
	if !serverExists(in.input) {
		return errors.New("Input server not known by debug server")
	}
	if !serverExists(in.output) {
		return errors.New("Output server not known by debug server")
	}

	addRuleToServer(in, in.input, in.output)

	*out = true
	return nil
}

func (t *Check) GetRule(in ServerConnection, out *Behavior) error {
	if !serverExists(in.Input) {
		return errors.New("Input server not known by debug server")
	}
	if !serverExists(in.Output) {
		return errors.New("Output server not known by debug server")
	}

	out = rules[in]

	return nil
}

/* TODO Make debugRpcServer runner
Additional imports needed:
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"


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
}*/
