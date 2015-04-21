package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func commitOneValue(t *testing.T, value string, resultChan chan bool,
	cfg *config.Config) {
	// Initialize client
	client := raftClient.CreateRaftClient(cfg)

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
	var testValues = []string{"a", "b", "c", "d", "e"}
	var numClients = len(testValues)
	var readValues []string

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

	resultChan := make(chan bool, numClients)
	for _, testVal := range testValues {
		go commitOneValue(t, testVal, resultChan, &cfg)
	}

	// Initialize client
	client := raftClient.CreateRaftClient(&cfg)

	for i := 1; i <= numClients; i++ {
		if res := <-resultChan; res == false {
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

	// Sleep for 500 ms
	time.Sleep(500 * time.Millisecond)

	// Verify that the log has been properly replicated
	for _, server := range serverNames {
		for i, testVal := range readValues {
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

}
