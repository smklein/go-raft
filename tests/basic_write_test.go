package tests

import (
	"config"
	"raftClient"
	"testing"
	"time"
)

func TestBasicWrite(t *testing.T) {
	var cfg config.Config
	//var serverNames []string
	var testValue = "testValue"

	t.Logf("Basic write test started")

	if !config.LoadConfig(&cfg) {
		t.Errorf("Cannot load config")
	}

	//serverNames = cfg.GetServerNames()

	//TODO: Initialize all servers

	// Initialize client
	client := raftClient.RaftClient{}

	// Commit a single value
	err := client.Commit(testValue)
	if err != nil {
		t.Errorf("Commit failure: %s", err)
		return
	}

	// Sleep for 100 ms
	time.Sleep(100 * time.Millisecond)

	// Verify that the log has been properly replicated
	readValue, err := client.ReadLog(0)
	if err != nil {
		t.Errorf("Log reading failure: %s", err)
		return
	}

	if readValue != testValue {
		t.Errorf("Incorrect value read - testValue: <%s> - readValue: <%s>", testValue, readValue)
		return
	}


	t.Errorf("Failed")
}