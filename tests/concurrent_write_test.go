package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func commitOneValue(t * testing.T, value string, resultChan chan bool) {
	// Initialize client
	client := raftClient.RaftClient{}

	// Commit a single value
	err := client.Commit(value)
	if err != nil {
		t.Errorf("Commit failure: %s", err)
		resultChan <- false
		return
	}
	resultChan <- true
}

func TestConcurrentWrite(t *testing.T) {
	var cfg config.Config
	var serverNames []string
	var testValues = []string{"", "", "", "", ""}
	var numClients = len(testValues)
	var readValues = make([]string, numClients)

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

	resultChan := make(chan bool, numClients)
	for _, testVal := range(testValues) {
		go commitOneValue(t, testVal, resultChan)
	}

	// Initialize client
	client := raftClient.RaftClient{}

	for i := 0; i < numClients; i++ {
		if <-resultChan == false {
			t.Errorf("Client write failure, terminating test")
			return
		} else {
			readValue, err := client.ReadLog(i)
			if err != nil {
				t.Errorf("Log reading failure at index %d: %s", i, err)
				return
			}
			readValues = append(readValues, readValue)
		}
	}
	
	t.Logf("Values written and read from system")

	// Sleep for 100 ms
	time.Sleep(100 * time.Millisecond)

	// Verify that the log has been properly replicated
	for _, server := range(serverNames) {
		for i, testVal := range(readValues) {
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