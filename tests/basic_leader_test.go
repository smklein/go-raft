package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func TestBasicLeader(t *testing.T) {
	var cfg config.Config
	var serverNames []string

	t.Logf("Basic leader test started")

	if !config.LoadConfig(&cfg) {
		t.Errorf("Cannot load config")
	}

	serverNames = cfg.GetServerNames()
	if serverNames == nil {
		return
	}

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
	err = sm.StartAllServers()
	if err != nil {
		t.Errorf("Failure starting raft servers: %s", err)
		return
	}
	t.Logf("All servers started")
	time.Sleep(5000 * time.Millisecond)

	// Initialize client
	client := raftClient.CreateRaftClient(&cfg)
	if client == nil {
		t.Errorf("Could not start client")
		return
	}
	time.Sleep(100 * time.Millisecond)

	// Verify that the log has been updated
	_, err = client.ReadLog(1)
	if err == nil {
		t.Errorf("Log reading failure -- expected empty log.")
		return
	}

	numLeaders := 0
	numFollowers := 0
	// Verify that there is one leader.
	for _, server := range serverNames {
		serverState, err := client.DebugGetServerStatus(server)
		if err != nil {
			t.Errorf("Server status failure at server <%s>: %s", server, err)
			return
		}
		if serverState == "leader" {
			numLeaders += 1
		} else if serverState == "follower" {
			numFollowers += 1
		} else if serverState == "candidate" {
			t.Errorf("Invalid server state: %s, ", serverState)
			return
		} else {
			t.Errorf("Invalid server state: %s, ", serverState)
			return
		}
	}

	if numLeaders != 1 {
		t.Errorf("Expected one leader, got %d leaders", numLeaders)
		return
	}
	if numFollowers != len(serverNames)-1 {
		t.Errorf("Expected %d followers, got %d followers", len(serverNames)-1,
			numFollowers)
		return
	}
	t.Logf("Test passed")
	return
}
