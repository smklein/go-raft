package raftClient

import (
	"config"
	"errors"
	"fmt"
	"net/rpc"
	"strconv"
)

type RaftClient struct {
	leader      string
	connections map[string]*rpc.Client
}

func CreateRaftClient(cfg *config.Config) *RaftClient {
	client := &RaftClient{}
	client.connections = make(map[string]*rpc.Client)
	fmt.Println(cfg.Servers)
	for _, server := range cfg.Servers {
		if server.Name != "debugServer" {
			addr := server.Address + ":" + strconv.Itoa(server.Port)
			con, err := rpc.DialHTTP("tcp", addr)
			if err != nil {
				fmt.Printf("Failed to connect with server <%s> when starting client\n", server.Name)
				fmt.Println(err)
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
	fmt.Println("[RAFT CLIENT] ReadLog called for index: ", index)
	fmt.Println(client.connections)
	var err error
	for server, conn := range client.connections {
		fmt.Println("[RAFT CLIENT] Observing server: ", server)
		var value string
		err = conn.Call("RaftServer.ReadLog", index, &value)
		if err == nil {
			fmt.Println("Read log at index: ", index, value)
			return value, err
		}
	}
	return "", err
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
	err := conn.Call("RaftServer.GetServerStatus", 0, &value)
	if err != nil {
		return "", err
	}
	return value, err
}

func (client *RaftClient) DebugGetRaftLeader() (string, error) {
	var leader string
	for server, _ := range client.connections {
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
	return errors.New("Unimplemented")
}
