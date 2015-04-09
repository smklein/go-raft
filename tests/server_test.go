package tests

import (
	"server"
	"serverManagement"
	"testing"
)

func TestBasic(t *testing.T) {
	if !server.GimmeTrue() {
		t.Errorf("Failed")
	}
}

func TestServerStartup(t *testing.T) {
	t.Logf("Test Server Startup")
	sm := serverManagement.StartAllServers()
	t.Logf("All servers started")
	if err := sm.KillServer("server1"); err != nil {
		t.Log(err)
		t.Errorf("Failed")
	}
	if err := sm.RestartServer("server1"); err != nil {
		t.Log(err)
		t.Errorf("Failed")
	}
}
