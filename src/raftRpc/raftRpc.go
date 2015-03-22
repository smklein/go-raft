package raftRpc

import (
	"net/rpc"
	"os"
)

var (
	DEBUG         = true
	serverAddress = "localhost:8081"
)

type RaftClient struct {
	client      *rpc.Client
	debugClient *rpc.Client
}

func DialHTTP(network, address string) (*RaftClient, error) {
	c, err := rpc.DialHTTP(network, address)
	if err != nil {
		return nil, err
	}

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

func (rc *RaftClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if DEBUG {
		err := rc.debugClient.Call(serviceMethod, args, reply)
		if err != nil {
			// If DEBUG is on, we NEED the ability to contact the debug server.
			os.Exit(-1)
		}
		// TODO React to debug server.
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
