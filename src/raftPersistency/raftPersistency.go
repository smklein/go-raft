package raftPersistency

import (
	"config"
	"encoding/gob"
	"errors"
	"fmt"
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
	path := getDataPath(server)
    dataFile, err := os.Open(path)
    if err != nil {
    	return nil, err
    }
    dataDecoder := gob.NewDecoder(dataFile)
    state := &RaftPersistentState{}
    err = dataDecoder.Decode(&state)
    if err != nil {
    	return state, nil
    } else {
    	return state, dataFile.Close()
    }
}

func (state *RaftPersistentState) SaveState(server string) error {
	path := getDataPath(server)
    dataFile, err := os.Create(path)
    if err != nil {
    	return err
    }
    dataEncoder := gob.NewEncoder(dataFile)
    dataEncoder.Encode(state)
    return dataFile.Close()
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
