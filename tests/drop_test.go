package tests

import (
	"config"
	"debugRpcServer"
	"net/rpc"
	"serverManagement"
	"strconv"
	"testing"
)

var cfg config.Config
var serverNames []string
var debugServer *rpc.Client

func updateServer(t *testing.T, s1 string, drop bool) bool {
	for _, s2 := range serverNames {
		if s1 != s2 {
			var result bool
			args := debugRpcServer.RuleCommandRpcInput{s1, s2, "drop", drop}
			err := debugServer.Call("Check.AddRule", args, &result)
			if err != nil {
				t.Errorf("Error at debug server for %s (in) and %s (out): %s", s1, s2, err)
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
	err := serverManagement.StartDebugServer()
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	sm := serverManagement.StartAllServers()
	if sm == nil {
		t.Errorf("Could not start servers")
	}
	if !config.LoadConfig(&cfg) {
		t.Errorf("Cannot load config")
	}
	serverNames = make([]string, len(cfg.Servers)-1)
	i := 0
	for _, server := range cfg.Servers {
		if server.Name != "debugServer" {
			serverNames[i] = server.Name
			i++
		} else {
			debug, err := rpc.DialHTTP("tcp", server.Address+":"+strconv.Itoa(server.Port))
			if err != nil {
				t.Errorf("Cannot contact debug server:\n%s", err)
				return
			}
			debugServer = debug
		}
	}
	for _, server1 := range serverNames {
		t.Logf("Dropping first server: %s", server1)
		result := updateServer(t, server1, true)
		if !result {
			return
		}

		// DO some shit - all logs up to date except server 1
		// 1. Write value to system
		// 2. Verify values on all servers except 1

		for _, server2 := range serverNames {
			if server2 == server1 {
				continue
			}
			t.Logf("Dropping second server: %s", server2)
			result := updateServer(t, server2, true)
			if !result {
				return
			}

			// Do some shit - all logs up to date except servers 1 and 2
			// 1. Write value to system
			// 2. Verify values on all servers except 1 and 2

			for _, server3 := range serverNames {
				if server3 == server1 || server3 == server2 {
					continue
				}
				t.Logf("Dropping third server: %s", server3)
				result := updateServer(t, server3, true)
				if !result {
					return
				}

				// Do some shit - no progress here, client RPC should hang

				result = updateServer(t, server2, false)
				if !result {
					return
				}

				// Wait on the previous channel until it succeeds
			}

			result = updateServer(t, server1, false)
			if !result {
				return
			}

			// Log should be up to date for all servers except server 1
		}
		result = updateServer(t, server1, false)
		if !result {
			return
		}

		// Logs should all be equal and up to date

	}
	t.Logf("Test drop finished")
	//t.Error()
}
