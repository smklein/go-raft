package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func TestIllegalReadLog(t *testing.T) {
	var cfg config.Config
	var serverNames []string
	var testValue = "illegal read log test value"

	t.Logf("Illegal read log test started")

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
	err = sm.StartAllServers()
	if err != nil {
		t.Errorf("Failure starting raft servers")
		return
	}
	t.Logf("All servers started")
	time.Sleep(5000 * time.Millisecond)

	// Initialize client
	client := raftClient.CreateRaftClient(&cfg)

	// Commit a single value
	err = client.Commit(testValue)
	if err != nil {
		t.Errorf("Commit failure: %s", err)
		return
	}
	t.Logf("Value written")

	time.Sleep(500 * time.Millisecond)

	// Verify that the log has been updated
	readValue, err := client.ReadLog(1)
	if err != nil {
		t.Errorf("Log reading failure: %s", err)
		return
	}
	t.Logf("Value read")

	if readValue != testValue {
		t.Errorf("Incorrect value read - testValue: <%s> - readValue: <%s>", testValue, readValue)
		return
	}

	// Verify that the log has been properly replicated
	for _, server := range serverNames {
		readValue, err = client.ReadLogAtServer(1, server)
		if err != nil {
			t.Errorf("Log reading failure at server <%s>: %s", server, err)
			return
		}
		if readValue != testValue {
			t.Errorf("Incorrect value read at server <%s> - testValue: <%s> - readValue: <%s>", server, testValue, readValue)
			return
		}
	}

	_, err = client.ReadLog(2)
	if err == nil {
		t.Errorf("Expected client to fail when reading log value 2")
	} else if err.Error() != "Index not applied to state machine" {
		t.Errorf("Expected a different error value (TODO errval) ")
	}

	t.Logf("Test passed")
}
