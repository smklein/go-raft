package raftClient

import (
	"config"
	"net/rpc"
	"errors"
	"fmt"
	"strconv"
)

type RaftClient struct {
	leader string
	connections map[string]*rpc.Client
}

func CreateRaftClient(cfg *config.Config) *RaftClient{
	client := &RaftClient{}
	client.connections = make(map[string]*rpc.Client)
	for _, server := range(cfg.Servers) {
		if server.Name != "debugServer" {
			addr := server.Address + ":" + strconv.Itoa(server.Port)
			con, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				fmt.Println("Failed to connect with server <%s> when starting client", server.Name)
				return nil
			}
			client.connections[server.Name] = con
		}
	}
	return client
}

func (client *RaftClient) Commit(value string) error {
	if client.leader == "" {
		leader, err := client.DebugGetRaftLeader()
		if err != nil {
			return err
		}
		client.leader = leader
	}
	var result string
	err := client.connections[client.leader].Call("RaftServer.Commit", value, &result)
	if err != nil {
		return err
	}
	//TODO: proccess info in result
	return nil
}

func (client *RaftClient) ReadLog(index int) (string, error) {
	var finalErr error
	for server, conn := range(client.connections) {
		var value string
		err := conn.Call("RaftServer.ReadLog", index, &value)
		if err == nil {
			return value, err
		} else if server == client.leader {
			finalErr = err
		}
	}
	return "", finalErr
}

func (client *RaftClient) ReadLogAtServer(index int, server string) (string, error) {
	conn := client.connections[server]
	if conn == nil {
		return "", errors.New("Invalid server name")
	}
	var value string
	err := conn.Call("RaftServer.ReadLog", index, &value)
	if err != nil {
		return "", err
	}
	return value, err
}

func (client *RaftClient) ReadEntireLogAtServer(server string) ([]string, error) {
	conn := client.connections[server]
	if conn == nil {
		return nil, errors.New("Invalid server name")
	}
	value := make([]string, 0)
	err := conn.Call("RaftServer.ReadLog", nil, &value)
	if err != nil {
		return nil, err
	}
	return value, err
}

/*
DEBUG function, asks if server is "leader", "follower", or "candidate"
*/
func (client *RaftClient) DebugGetServerStatus(server string) (string, error) {
	conn := client.connections[server]
	if conn == nil {
		return "", errors.New("Invalid server name")
	}
	var value string
	err := conn.Call("RaftServer.GetServerStatus", nil, &value)
	if err != nil {
		return "", err
	}
	return value, err
}

func (client *RaftClient) DebugGetRaftLeader() (string, error) {
	var leader string
	for server, _ := range(client.connections) {
		status, err := client.DebugGetServerStatus(server)
		if err != nil {
			return "", err
		}
		if leader == "" {
			if status == "leader" {
				leader = server
			}
		} else if status == "leader" {
			return "", errors.New(fmt.Sprintf("Multiple leaders detected: <%s>, <%s>", leader, server))
		}
	}
	if leader == "" {
		return "", errors.New("No leader found")
	} else {
		return leader, nil
	}
}

func (client *RaftClient) DebugCommitToServer(value, server string) error {
	return nil
}