package tests

import (
	"config"
	"raftClient"
	"raftPersistency"
	"serverManagement"
	"testing"
	"time"
)

func TestLeaderHandoff(t *testing.T) {
	var cfg config.Config
	var serverNames []string
	var testValue = "Some test value"

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
	firstLeader, err := client.DebugGetRaftLeader()
	if err != nil {
		t.Errorf("Could not determine initial leader: %s", err)
		return
	}
	//TODO: block communication to/from old leader

	// Sleep for 100 ms
	time.Sleep(100 * time.Millisecond)

	// Verify that a new leader has been elected
	var newLeader string
	for _, server := range(serverNames) {
		if server != firstLeader {
			status, err := client.DebugGetServerStatus(server)
			if err != nil {
				t.Errorf("Error contacting server <%s>: %s", server, err)
				return
			}
			if newLeader == "" {
				if status == "leader" {
					newLeader = server
				}
			} else if status == "leader" {
				t.Errorf("Multiple leaders detected: <%s>, <%s>", newLeader, server)
				return
			}
		}
	}
	if newLeader == "" {
		t.Errorf("No new leader found")
		return
	}

	// Verify that the old leader still thinks it is a leader
	firstLeaderStatus, err := client.DebugGetServerStatus(firstLeader)
	if firstLeaderStatus != "leader" {
		t.Errorf("First leader's status changed to %s", firstLeaderStatus)
		return
	}

	// Send commit to new leader and verify failure
	err = client.DebugCommitToServer(testValue, newLeader)
	if err != nil {
		t.Errorf("Unexpected error commiting at new leader <%s>:", newLeader, err)
		return
	}

	// Send commit to old leader and verify success
	err = client.DebugCommitToServer(testValue, firstLeader)
	if err == nil {
		t.Errorf("Unexpected successful commit at first leader <%s>", firstLeader)
		return
	} else if err.Error() != "Not leader" {
		t.Errorf("Unexpected error at first leader <%s>: %s", firstLeader, err)
	}

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
		if server != firstLeader {
			if err != nil {
				t.Errorf("Log reading failure at server <%s>: %s", server, err)
				return
			}
			if readValue != testValue {
				t.Errorf("Incorrect value read at server <%s> - testValue: <%s> - readValue: <%s>", server, testValue, readValue)
				return
			}
		} else {
			if err == nil {
				t.Errorf("Unexected success reading log at first leader <%s> with value %s", firstLeader, readValue)
			} else if err.Error() != "Log does not contain index" {
				t.Errorf("Expected a different error value at first leader <%s>", firstLeader)
			}
		}
	}

	//TODO: open communication to/from old leader

	// Sleep for 100 ms
	time.Sleep(100 * time.Millisecond)

	//Verify that the first leader is a follower
	firstLeaderStatus, err = client.DebugGetServerStatus(firstLeader)
	if firstLeaderStatus != "follower" {
		t.Errorf("First leader's status expected to be <follower> is <%s>", firstLeaderStatus)
		return
	}

	//Verify that the log has been updated at the first leader
	readValue, err = client.ReadLogAtServer(0, firstLeader)
	if err != nil {
		t.Errorf("Log reading failure at firstLeader <%s>: %s", firstLeader, err)
		return
	}
	if readValue != testValue {
		t.Errorf("Incorrect value read at first leader <%s> - testValue: <%s> - readValue: <%s>", firstLeader, testValue, readValue)
		return
	}
}