package tests

import (
	"config"
	"debugRpcServer"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func TestNonMajorityElection(t *testing.T) {
	var cfg config.Config
	var serverNames []string

	t.Logf("Non majority election test started")

	if !config.LoadConfig(&cfg) {
		t.Errorf("Cannot load config")
	}

	serverNames = cfg.GetServerNames()
	err := raftPersistency.DeleteAllLogs()
	if err != nil {
		t.Errorf("Could not clear log files: %s", err)
		return
	}

	err = serverManagement.StartDebugServer()
	if err != nil {
		t.Errorf("Failure starting debug server: %s", err)
		return
	}
	addr := cfg.GetDebugAddr()
	dbg := debugRpcServer.CreateDebugServerConnection(addr, serverNames)
	// Server 1 + 2 are in a "group"
	dbg.AddRule("server1", "server3", "drop", true)
	dbg.AddRule("server1", "server4", "drop", true)
	dbg.AddRule("server1", "server5", "drop", true)
	dbg.AddRule("server2", "server3", "drop", true)
	dbg.AddRule("server2", "server4", "drop", true)
	dbg.AddRule("server2", "server5", "drop", true)

	// Server 3 + 4 are in a "group"
	dbg.AddRule("server3", "server1", "drop", true)
	dbg.AddRule("server3", "server2", "drop", true)
	dbg.AddRule("server3", "server5", "drop", true)
	dbg.AddRule("server4", "server1", "drop", true)
	dbg.AddRule("server4", "server2", "drop", true)
	dbg.AddRule("server4", "server5", "drop", true)

	// Server 5 cannot talk to anyone.
	dbg.AddRule("server5", "server1", "drop", true)
	dbg.AddRule("server5", "server2", "drop", true)
	dbg.AddRule("server5", "server3", "drop", true)
	dbg.AddRule("server5", "server4", "drop", true)

	sm := serverManagement.StartAllServers()
	if sm == nil {
		t.Errorf("Failure starting raft server")
		return
	}
	t.Logf("All servers started")

	client := raftClient.RaftClient{}
	time.Sleep(250 * time.Millisecond)
	for _, s := range serverNames {
		status, err := client.DebugGetServerStatus(s)
		if err != nil {
			t.Errorf("Server status failure: %s", err)
		}
		if status != "candidate" {
			t.Errorf("Server %s thinks it is a %s", s, status)
		}
	}

	err = client.Commit("asdf")
	if err == nil {
		t.Errorf("Expected committing to return an error")
	}
	if err.Error() != "Timed out" {
		t.Errorf("Expected committing to time out")
	}

	t.Logf("Test passed")

	sm.KillAllServers()
}
