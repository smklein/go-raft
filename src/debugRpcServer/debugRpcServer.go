package debugRpcServer

import (
	"config"
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

var rules map[ServerConnection]*Behavior = make(map[ServerConnection]*Behavior)
var Cf config.Config

func (t *Check) CheckRPC(in string, out *string) error {
	s := []string{in, "Delay"}
	fmt.Println("[Debug server] Sees input: ", in)
	*out = strings.Join(s, " ")
	return nil
}

type RuleCommandRpcInput struct {
	Input    string
	Output   string
	Behavior string
	On       bool
}

func serverExists(name string) bool {
	for _, s := range Cf.Servers {
		if s.Name == name {
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

	if in.Behavior == "delay" {
		behavior.Delay = in.On
	}
	if in.Behavior == "drop" {
		behavior.Drop = in.On
	}

	rules[sc] = behavior
}

func (t *Check) AddRule(in RuleCommandRpcInput, out *bool) error {
	if !serverExists(in.Input) {
		return errors.New("Input server not known by debug server")
	}
	if !serverExists(in.Output) {
		return errors.New("Output server not known by debug server")
	}

	addRuleToServer(in, in.Input, in.Output)

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
