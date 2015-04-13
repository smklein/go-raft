package tests

import (
	"config"
	"debugRpcServer"
	"serverManagement"
	"strconv"
	"testing"
)

func TestDropServers(t *testing.T) {
	var cfg config.Config
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
	serverNames := make([]string, len(cfg.Servers)-1)
	i := 0
	var addr string
	for _, server := range cfg.Servers {
		if server.Name != "debugServer" {
			serverNames[i] = server.Name
			i++
		} else {
			addr = server.Address + ":" + strconv.Itoa(server.Port)
		}
	}
	dbg := debugRpcServer.CreateDebugServerConnection(addr, serverNames)
	if dbg == nil {
		t.Errorf("Could not dial debug server")
	}

	for _, server1 := range serverNames {
		t.Logf("Dropping first server: %s", server1)
		if err := dbg.AddRuleIncomingFromServer(server1, "drop", true); err != nil {
			t.Errorf("Could not add incoming rules")
		}
		if err := dbg.AddRuleOutputToServer(server1, "drop", true); err != nil {
			t.Errorf("Could not add outgoing rule")
		}

		// DO some shit - all logs up to date except server 1
		// 1. Write value to system
		// 2. Verify values on all servers except 1

		for _, server2 := range serverNames {
			if server2 == server1 {
				continue
			}
			t.Logf("Dropping second server: %s", server2)
			dbg.AddRuleIncomingFromServer(server2, "drop", true)
			dbg.AddRuleOutputToServer(server2, "drop", true)

			// Do some shit - all logs up to date except servers 1 and 2
			// 1. Write value to system
			// 2. Verify values on all servers except 1 and 2

			for _, server3 := range serverNames {
				if server3 == server1 || server3 == server2 {
					continue
				}
				t.Logf("Dropping third server: %s", server3)
				dbg.AddRuleIncomingFromServer(server3, "drop", true)
				dbg.AddRuleOutputToServer(server3, "drop", true)

				// Do some shit - no progress here, client RPC should hang

				dbg.AddRuleIncomingFromServer(server2, "drop", false)
				dbg.AddRuleOutputToServer(server2, "drop", false)

				// Wait on the previous channel until it succeeds
			}

			dbg.AddRuleIncomingFromServer(server1, "drop", false)
			dbg.AddRuleOutputToServer(server1, "drop", false)

		}

		// Logs should all be equal and up to date

	}
	t.Logf("Test drop finished")
	//t.Error()
}
