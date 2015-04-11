package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func TestSequentialWrite(t *testing.T) {
	var cfg config.Config
	var serverNames []string
	var testValues = []string{"", "", "", "", ""}

	t.Logf("Sequential write test started")

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

	// Commit the sequence of values value
	for i, testVal := range(testValues) {
		err = client.Commit(testVal)
		if err != nil {
			t.Errorf("Commit failure on string #%d <%s>: %s", i, testVal, err)
			return
		}
	}
	t.Logf("Values written")

	// Sleep for 100 ms
	time.Sleep(100 * time.Millisecond)

	// Verify that the log has been updated
	for i, testVal := range(testValues) {
		readValue, err := client.ReadLog(i)
		if err != nil {
			t.Errorf("Log reading failure on string #%d <%s>: %s", i, testVal, err)
			return
		}
		if readValue != testVal {
			t.Errorf("Incorrect value read - index: <%d> - testValue: <%s> - readValue: <%s>", i, testVal, readValue)
			return
		}
	}
	t.Logf("Values read and verified in system")

	// Verify that the log has been properly replicated
	for _, server := range(serverNames) {
		for i, testVal := range(testValues) {
			readValue, err := client.ReadLogAtServer(i, server)
			if err != nil {
				t.Errorf("Log reading for string #%d failure at server <%s>: %s", i, server, err)
				return
			}
			if readValue != testVal {
				t.Errorf("Incorrect value read at server <%s> - index: <%d> - testValue: <%s> - readValue: <%s>", server, i, testVal, readValue)
				return
			}
		}
	}
	t.Logf("Test passed")

	sm.KillAllServers()
}