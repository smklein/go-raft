package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func TestBasicWrite(t *testing.T) {
	var cfg config.Config
	var serverNames []string
	var testValue = "basic write test value"

	t.Logf("Basic write test started")

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

	time.Sleep(1000 * time.Millisecond)

	// Verify that the log has been updated
	readValue, err := client.ReadLog(1)
	if err != nil {
		t.Errorf("Log reading failure: %s", err)
		return
	}
	t.Logf("Value read correctly from generic ReadLog.")

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
		t.Logf("Server <%s> read the value correctly\n", server)
	}
	readValue, err = client.ReadLog(2)
	if err == nil {
		t.Errorf("Reading index at 2 expected error, got %s", readValue)
		return
	}
	t.Logf("Entry 2 is still empty")

	t.Logf("Test passed")
}
