package raftPersistency

import (
	"config"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"raftRpc"
)

/* Persistent raft state on all servers */
type RaftPersistentState struct {
	CurrentTerm int
	VotedFor    int // -1 represents null.
	Log         []raftRpc.LogEntry
}

func LoadState(server string) (*RaftPersistentState, error) {
	return &RaftPersistentState{}, errors.New("No state found")
}

func (server string, state *RaftPersistentState) SaveState() error {
	/* b, err := json.Marshal(state)
    if err != nil {
        return err
    }
    err = ioutil.WriteFile(getDataPath(name, d1, 0644) */
   	return nil
}

/**
THESE FUNCTIONS ARE NOT TO BE RUN WHILE A SERVER IS ACTIVE.
*/

func getDataPath(serverName string) string {
	return os.Getenv("GOPATH") + "data/" + serverName
}

func DeleteLog(serverName string) error {
	return os.Truncate(getDataPath(serverName), 0)
}

func DeleteAllLogs() error {
	var Cf config.Config
	if !config.LoadConfig(&Cf) {
		fmt.Println("[RAFT PERSISTENCY] Could not load config")
		return errors.New("Cannot load config")
	}
	for _, s := range Cf.Servers {
		if s.Name != "debugServer" {
			err := DeleteLog(s.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
