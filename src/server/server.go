package server

import (
	"config"
	"fmt"
	//"raftRpc"
)

type RaftServer int

var Cf config.Config

type RaftServerInput struct {
	placeholder string
}

type RaftServerOutput struct {
	placeholder string
}

func (rs *RaftServer) Commit(in RaftServerInput, out *RaftServerOutput) error {
	fmt.Println("[RAFT SERVER] Commit Called")
	return nil
}

func GimmeTrue() bool {
	return true
}
