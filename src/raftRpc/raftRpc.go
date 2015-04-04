package raftRpc

import (
	"debugRpcServer"
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"time"
)

var (
	DEBUG         = true
	serverAddress = ":8081" // TODO Read this from config.yaml
)

type RaftClient struct {
	client      *rpc.Client
	debugClient *rpc.Client
}

type RaftClientArgs struct {
	// TODO Fill this!
	inputServer  string
	outputServer string
}

type RaftClientReply struct {
	// TODO Fill this!
}

func DialHTTP(network, address string) (*RaftClient, error) {
	fmt.Println("Dial HTTP called")
	c, err := rpc.DialHTTP(network, address)
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully dialed: ", address)

	var d *rpc.Client
	if DEBUG {
		d, err = rpc.DialHTTP(network, serverAddress)
		if err != nil {
			// If DEBUG is on, we NEED the ability to contact the debug server.
			os.Exit(-1)
		}
	}
	return &RaftClient{client: c, debugClient: d}, err
}

func (rc *RaftClient) Call(serviceMethod string, args RaftClientArgs, reply *RaftClientReply) error {
	if DEBUG {
		sc := debugRpcServer.ServerConnection{args.inputServer, args.outputServer}
		var behavior debugRpcServer.Behavior
		err := rc.debugClient.Call("GetRule", sc, &behavior)
		if err != nil {
			// If DEBUG is on, we NEED the ability to contact the debug server.
			os.Exit(-1)
		}
		if behavior.Drop {
			fmt.Println("Applying DROP rule")
			return errors.New("Message dropped by RPC layer")
		} else if behavior.Delay {
			fmt.Println("Applying DELAY rule")
			time.Sleep(2000 * time.Millisecond) // TODO Add ability to wait for different times.
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
