package tests

import (
	"server"
	"testing"
)

func TestBasic(t *testing.T) {
	t.Logf("Test basic started")
	if !server.GimmeTrue() {
		t.Errorf("Failed")
	}
}
