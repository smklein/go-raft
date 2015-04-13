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

func TestConcurrentLeaderElection(t *testing.T) {
	var cfg config.Config
	var serverNames []string

	t.Logf("Concurrent Leader Election test started")

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
	// Servers cannot talk to anyone.
	dbg.AddRuleIncomingFromServer("server1", "drop", true)
	dbg.AddRuleIncomingFromServer("server2", "drop", true)
	dbg.AddRuleIncomingFromServer("server3", "drop", true)
	dbg.AddRuleIncomingFromServer("server4", "drop", true)
	dbg.AddRuleIncomingFromServer("server5", "drop", true)

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

	// Servers can talk! RACE TO BECOME LEADER!
	dbg.AddRuleIncomingFromServer("server1", "drop", false)
	dbg.AddRuleIncomingFromServer("server2", "drop", false)
	dbg.AddRuleIncomingFromServer("server3", "drop", false)
	dbg.AddRuleIncomingFromServer("server4", "drop", false)
	dbg.AddRuleIncomingFromServer("server5", "drop", false)
	time.Sleep(250 * time.Millisecond)

	leaderCount := 0
	followerCount := 0
	for _, s := range serverNames {
		status, err := client.DebugGetServerStatus(s)
		if err != nil {
			t.Errorf("Server status failure: %s", err)
		}
		if status == "leader" {
			leaderCount += 1
		} else if status == "follower" {
			followerCount += 1
		} else {
			t.Errorf("Server %s thinks it is a %s", s, status)
		}
	}

	if leaderCount != 1 {
		t.Errorf("System has %d leaders", leaderCount)
	}
	t.Logf("Test passed")
	sm.KillAllServers()
}
