package tests

import (
	"raftPersistency"
	"raftRpc"
	"testing"
)

func TestLeaderHandoff(t *testing.T) {
	testServer := "server1"
	testTerm := 5
	testVotedFor := -1
	testCommand := "test command"
	t.Logf("Persistent state test started")
	raftPersistency.DeleteLog(testServer)
	state, err := raftPersistency.LoadState(testServer)
	if err != nil {
		t.Errorf("Error clearing log for server <%s>: %s", testServer, err)
	}

	// Modify state
	state.CurrentTerm = testTerm
	state.VotedFor = testVotedFor
	state.Log = append(state.Log, raftRpc.LogEntry{testCommand, testTerm})
	
	t.Logf("%s", state)
	// Write state to file
	err = state.SaveState(testServer)
	if err != nil {
		t.Errorf("Error saving state for server <%s>: %s", testServer, err)
	}

	state, err = raftPersistency.LoadState(testServer)
	if err != nil {
		t.Errorf("Error loading state for server <%s>: %s", testServer, err)
	}

	if state.CurrentTerm != testTerm {
		t.Errorf("Error reading term - expected <%d> but found <%d>: %s", testTerm, state.CurrentTerm, err)
	}

	if state.VotedFor != testVotedFor{
		t.Errorf("Error reading voted for - expected <%d> but found <%d>: %s", testVotedFor, state.VotedFor, err)
	}

	if state.Log[0].Command != testCommand {
		t.Errorf("Error reading log command - expected <%s> but found <%s>: %s", testCommand, state.Log[0].Command, err)
	}

	if state.Log[0].Term != testTerm {
		t.Errorf("Error reading log term - expected <%d> but found <%d>: %s", testTerm, state.Log[0].Term, err)
	}

	t.Logf("Test passed")
}