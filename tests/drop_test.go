package tests

import (
	"config"
	"debugRpcServer"
	"net/rpc"
	"strconv"
	"testing"
)

var cfg config.Config
var serverNames []string
var debugServer *rpc.Client

func updateServer(t *testing.T, s1 string, drop bool) bool{
	for _, s2 := range(serverNames) {
		if s1 != s2 {
			var result bool
			args := debugRpcServer.RuleCommandRpcInput{s1, s2, "drop", drop}
			err := debugServer.Call("Check.AddRule", args, &result)
			if err != nil {
				t.Errorf("Error at debug server for %s (in) and %s (out): %s" ,s1, s2, err)
				return false
			}
			if !result {
				t.Errorf("Debug server failed to update drop from %s to %s with %b", s1, s2, drop)
				return false
			}
		}
	}
	return true
}

func TestDropServers(t *testing.T) {
	t.Logf("Test drop started")
	if !config.LoadConfig(&cfg) {
		t.Errorf("Cannot load config")
	}
	serverNames = make([]string, len(cfg.Servers) - 1)
	i := 0
	for _, server := range(cfg.Servers) {
		if server.Name != "debugServer" {
			serverNames[i] = server.Name
			i++
		} else {
			debug, err := rpc.DialHTTP("tcp", server.Address + ":" + strconv.Itoa(server.Port))
			if err != nil {
				t.Errorf("Cannot contact debug server:\n%s", err)
				//return
			}
			debugServer = debug
		}
	}
	for _, server1 := range(serverNames) {
		t.Logf("Dropping first server: %s", server1)
		result := updateServer(t, server1, true)
		if !result {
			return
		}
		for _, server2 := range(serverNames) {
			if server2 == server1 {
				continue
			}
			t.Logf("Dropping second server: %s", server2)
			for _, server3 := range(serverNames) {
				if server3 == server1 || server3 == server2 {
					continue
				}
				t.Logf("Dropping third server: %s", server3)
			}
		}
		result = updateServer(t, server1, false)
		if !result {
			return
		}

	}
	t.Logf("Test drop finished")
	//t.Error()
}