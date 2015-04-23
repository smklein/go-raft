package raftRpc

import (
	"config"
	"debugRpcServer"
	"errors"
	"fmt"
	"net/rpc"
	"strconv"
	"time"
)

var (
	DEBUG = true
)

type RaftClient struct {
	client       *rpc.Client
	debugClient  *rpc.Client
	inputServer  string
	outputServer string
}

type LogEntry struct {
	Command string
	Term    int
}

type AppendEntriesInput struct {
	Term         int
	LeaderId     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []*LogEntry
	LeaderCommit int
}
type AppendEntriesOutput struct {
	Term    int
	Success bool
}

type RequestVoteInput struct {
	Term         int
	CandidateId  string
	LastLogIndex int
	LastLogTerm  int
}
type RequestVoteOutput struct {
	Term        int
	VoteGranted bool
}

type RaftClientArgs struct {
	AppendEntriesIn *AppendEntriesInput
	RequestVoteIn   *RequestVoteInput
}

type RaftClientReply struct {
	AppendEntriesOut *AppendEntriesOutput
	RequestVoteOut   *RequestVoteOutput
}

func DialHTTP(network, address string, in, out string) (*RaftClient, error) {
	fmt.Println("\t[RAFT RPC] Dial HTTP called")
	c, err := rpc.DialHTTP(network, address)
	if err != nil {
		return nil, err
	}
	fmt.Println("\t[RAFT RPC] Successfully dialed: ", address)

	var d *rpc.Client
	if DEBUG {
		var cf config.Config
		var serverAddress string
		config.LoadConfig(&cf)
		for _, s := range cf.Servers {
			if s.Name == "debugServer" {
				serverAddress = s.Address + ":" + strconv.Itoa(s.Port)
			}
		}

		d, err = rpc.DialHTTP(network, serverAddress)
		if err != nil {
			panic(err)
		}
	}
	return &RaftClient{client: c,
		debugClient:  d,
		inputServer:  in,
		outputServer: out}, err
}

func (rc *RaftClient) Call(serviceMethod string, args RaftClientArgs,
	reply *RaftClientReply) error {
	if DEBUG {
		sc := debugRpcServer.ServerConnection{rc.inputServer, rc.outputServer}
		var behavior debugRpcServer.Behavior
		err := rc.debugClient.Call("Check.GetRule", sc, &behavior)
		if err != nil {
			// If DEBUG is on, we NEED the ability to contact the debug server.
			panic(err)
		}
		if behavior.Drop {
			fmt.Println("Applying DROP rule")
			return errors.New("Message dropped by RPC layer")
		} else if behavior.Delay {
			fmt.Println("Applying DELAY rule")
			time.Sleep(2000 * time.Millisecond)
			// TODO Add ability to wait for different times.
		}
	}
	return rc.client.Call(serviceMethod, args, reply)
}

func (rc *RaftClient) Close() error {
	if DEBUG {
		// If we get an error here, we can't do anything about it...
		rc.debugClient.Close()
	}
	return rc.client.Close()
}
