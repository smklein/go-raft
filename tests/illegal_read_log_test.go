package tests

import (
	"config"
	"raftClient"
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

	err := serverManagement.StartDebugServer()
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

	_, err = client.ReadLog(1)
	if err == nil {
		t.Errorf("Expected client to fail when reading log value 1")
	} else if err.Error() != "Log does not contain index" {
		// TODO Verify the error value here
		t.Errorf("Expected a different error value (TODO errval) ")
	}

	t.Logf("Test passed")

	sm.KillAllServers()
}