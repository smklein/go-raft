package tests

import (
	"server"
	"testing"
)

var config server.Config
var serverNames []string

func TestDropServers(t *testing.T) {
	t.Logf("Test drop started")
	if !server.LoadConfig(&config) {
		t.Errorf("Failed")
	}
	serverNames = make([]string, len(config.Servers) - 1)
	i := 0
	for _, server := range(config.Servers) {
		if server.Name != "debugServer" {
			serverNames[i] = server.Name
			i++
		}
	}
	for _, server1 := range(serverNames) {
		t.Logf("Dropping first server: %s", server1)
		for _, server2 := range(serverNames) {
			if server2 == server1 {
				continue
			}
			t.Logf("Dropping second server: %s", server2)
			for _, server3 := range(serverNames) {
				if server3 == server1 || server3 == server2 {
					continue
				}
				t.Logf("Dropping third server: %s", server3)
			}
		}
	}
	t.Logf("Test drop finished")
	t.Error()
}