package raftPersistency

import (
	"config"
	"errors"
	"fmt"
	"os"
)

/**
THESE FUNCTIONS ARE NOT TO BE RUN WHILE A SERVER IS ACTIVE.
*/

func getDataPath(serverName string) string {
	return "data/" + serverName
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
		err := DeleteLog(s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
