package tests

import (
	"config"
	"raftClient"
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
	sm := serverManagement.StartAllServers()
	if sm == nil {
		t.Errorf("Failure starting raft servers")
		return
	}
	t.Logf("All servers started")

	// Initialize client
	client := raftClient.RaftClient{}

	// Sleep for 100 ms
	time.Sleep(100 * time.Millisecond)

	// Verify that the log has been updated
	_, err = client.ReadLog(0)
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
		} else {
			t.Errorf("Invalid server state: %s, ", serverState)
		}
	}

	if numLeaders != 1 {
		t.Errorf("Expected one leader, got %d leaders", numLeaders)
	}
	if numFollowers != len(serverNames)-1 {
		t.Errorf("Expected %d followers, got %d followers", len(serverNames)-1,
			numFollowers)
	}
	t.Logf("Test passed")

	sm.KillAllServers()
}
