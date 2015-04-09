package raftClient

import (
	//"net/rpc"
)

type RaftClient struct {
	test int
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