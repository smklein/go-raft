package raftClient

import (
//"net/rpc"
	"errors"
)

type RaftClient struct {
	test int
	serverNames []string
}

func (client *RaftClient) Commit(value string) error {
	return nil
}

func (client *RaftClient) ReadLog(index int) (string, error) {
	return "", nil
}

func (client *RaftClient) ReadLogAtServer(index int, server string) (string, error) {
	return "", nil
}

func (client *RaftClient) ReadEntireLogAtServer(index int, server string) ([]string, error) {
	return nil, nil
}

/*
DEBUG function, asks if server is "leader", "follower", or "candidate"
*/
func (client *RaftClient) DebugGetServerStatus(server string) (string, error) {
	return "", nil
}

func (client *RaftClient) DebugGetRaftLeader() (string, error) {
	var leader string
	for _, server := range(client.serverNames) {
		status, err := client.DebugGetServerStatus(server)
		if err != nil {
			return "", err
		}
		if leader == "" {
			if status == "leader" {
				leader = server
			}
		} else if status == "leader" {
			return "", errors.New("Multiple leaders detected")
		}
	}
	if leader == "" {
		return "", errors.New("No leader found")
	} else {
		return leader, nil
	}
}