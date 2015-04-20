package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func TestBasicPersistentLog(t *testing.T) {
	var cfg config.Config
	var serverNames []string
	var testValue = "basic persistent log test value"

	t.Logf("Basic persistent log test started")
	if err := raftPersistency.DeleteAllLogs(); err != nil {
		t.Errorf("Could not delete logs at start of test")
	}

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

	// Initialize client
	client := raftClient.RaftClient{}

	// Commit a single value
	err = client.Commit(testValue)
	if err != nil {
		t.Errorf("Commit failure: %s", err)
		return
	}
	t.Logf("Value written")

	// Sleep for 100 ms
	time.Sleep(100 * time.Millisecond)

	// Verify that the log has been updated
	readValue, err := client.ReadLog(0)
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
		readValue, err = client.ReadLogAtServer(0, server)
		if err != nil {
			t.Errorf("Log reading failure at server <%s>: %s", server, err)
			return
		}
		if readValue != testValue {
			t.Errorf("Incorrect value read at server <%s> - testValue: <%s> - readValue: <%s>", server, testValue, readValue)
			return
		}
	}

	sm.KillAllServers()
	time.Sleep(100 * time.Millisecond)
	sm.RestartAllServers()
	time.Sleep(100 * time.Millisecond)

	// Verify that the log has been updated
	readValue, err = client.ReadLog(0)
	if err != nil {
		t.Errorf("(restarted) Log reading failure: %s", err)
		return
	}
	t.Logf("(restarted) Value read")

	if readValue != testValue {
		t.Errorf("(restarted) Incorrect value read - testValue: <%s> - readValue: <%s>", testValue, readValue)
		return
	}

	// Verify that the log has been properly replicated
	for _, server := range serverNames {
		readValue, err = client.ReadLogAtServer(0, server)
		if err != nil {
			t.Errorf("(restarted) Log reading failure at server <%s>: %s", server, err)
			return
		}
		if readValue != testValue {
			t.Errorf("(restarted) Incorrect value read at server <%s> - testValue: <%s> - readValue: <%s>", server, testValue, readValue)
			return
		}
	}

	t.Logf("Test passed")
	sm.KillAllServers()
}
