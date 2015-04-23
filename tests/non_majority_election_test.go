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

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

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

	sm := &serverManagement.ServerManager{}
	defer sm.KillAllServers()
	err = sm.StartDebugServer()
	if err != nil {
		t.Errorf("Failure starting debug server: %s", err)
		return
	}
	// Debug server must come online...
	time.Sleep(2000 * time.Millisecond)
	addr := cfg.GetDebugAddr()
	dbg, err := debugRpcServer.CreateDebugServerConnection(addr, serverNames)
	if err != nil {
		t.Errorf("Could not create debug server: %s", err)
		return
	}
	// Server 1 + 2 are in a "group"
	err = dbg.AddRule("server1", "server3", "drop", true)
	chk(err)
	err = dbg.AddRule("server1", "server4", "drop", true)
	chk(err)
	err = dbg.AddRule("server1", "server5", "drop", true)
	chk(err)
	err = dbg.AddRule("server2", "server3", "drop", true)
	chk(err)
	err = dbg.AddRule("server2", "server4", "drop", true)
	chk(err)
	err = dbg.AddRule("server2", "server5", "drop", true)
	chk(err)

	// Server 3 + 4 are in a "group"
	err = dbg.AddRule("server3", "server1", "drop", true)
	chk(err)
	err = dbg.AddRule("server3", "server2", "drop", true)
	chk(err)
	err = dbg.AddRule("server3", "server5", "drop", true)
	chk(err)
	err = dbg.AddRule("server4", "server1", "drop", true)
	chk(err)
	err = dbg.AddRule("server4", "server2", "drop", true)
	chk(err)
	err = dbg.AddRule("server4", "server5", "drop", true)
	chk(err)

	// Server 5 cannot talk to anyone.
	err = dbg.AddRule("server5", "server1", "drop", true)
	chk(err)
	err = dbg.AddRule("server5", "server2", "drop", true)
	chk(err)
	err = dbg.AddRule("server5", "server3", "drop", true)
	chk(err)
	err = dbg.AddRule("server5", "server4", "drop", true)
	chk(err)

	err = sm.StartAllServers()
	if err != nil {
		t.Errorf("Failure starting raft server")
		return
	}
	t.Logf("All servers started")
	time.Sleep(4000 * time.Millisecond)

	client := raftClient.CreateRaftClient(&cfg)
	for _, s := range serverNames {
		status, err := client.DebugGetServerStatus(s)
		if err != nil {
			t.Errorf("Server status failure: %s", err)
			return
		}
		if status == "leader" {
			t.Errorf("Server %s thinks it is a %s", s, status)
			return
		}
	}

	err = client.Commit("asdf")
	if err == nil {
		t.Errorf("Expected committing to return an error")
		return
	}
	if err.Error() != "No leader found" {
		t.Errorf("Expected different failure, was [%s]", err.Error())
		return
	}

	t.Logf("Test passed")
}
