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
	var testValues = []string{"a", "b", "c", "d", "e"}

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

	// Commit the sequence of values value
	for i, testVal := range testValues {
		err = client.Commit(testVal)
		if err != nil {
			t.Errorf("Commit failure on string #%d <%s>: %s", i, testVal, err)
			return
		}
	}
	t.Logf("Values written")

	time.Sleep(1000 * time.Millisecond)

	// Verify that the log has been updated
	for i, testVal := range testValues {
		readValue, err := client.ReadLog(i + 1)
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
	for _, server := range serverNames {
		for i, testVal := range testValues {
			readValue, err := client.ReadLogAtServer(i+1, server)
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
