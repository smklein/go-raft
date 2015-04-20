package server

import (
	"config"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"raftPersistency"
	"raftRpc"
	"strconv"
	"sync"
	"time"
)

type ClientResponse struct {
	err   error
	index int
}

/* All state on all servers */
type RaftServer struct {
	serverName string
	cf         config.Config
	serverMap  map[string]*raftRpc.RaftClient

	pState      *raftPersistency.RaftPersistentState
	state       string /* "follower", "candidate", "leader" */
	commitIndex int
	lastApplied int
	// Following used by followers only
	attemptCandidate bool
	// Following are used by leaders only.
	nextIndex    map[string]int
	matchIndex   map[string]int
	clientInput  chan string
	clientOutput chan int
	rpcLock      *sync.Mutex
}

/* Input arguments from client --> raft server */
type RaftServerInput struct {
	placeholder string
}

/* Output from raft server --> client */
type RaftServerOutput struct {
	placeholder string
}

/* Function to initialize, start up raft server */
func CreateRaftServer(serverName string) *RaftServer {
	fmt.Printf("[SERVER] Creating server %s\n", serverName)
	rs := &RaftServer{}
	if !config.LoadConfig(&rs.cf) {
		fmt.Printf("[SERVER] [ERROR] Could not load config\n")
		return nil
	}
	port := 0
	for _, s := range rs.cf.Servers {
		if s.Name == serverName {
			port = s.Port
		}
	}
	if port == 0 {
		fmt.Println("[SERVER] [ERROR] Port == 0")
		return nil
	}
	// TODO Store/Load!
	rs.pState = &raftPersistency.RaftPersistentState{}
	rs.pState.VotedFor = ""
	rs.pState.Log = []*raftRpc.LogEntry{&raftRpc.LogEntry{Command: "", Term: 0}}
	rs.pState.StateMachineLog = []*raftRpc.LogEntry{&raftRpc.LogEntry{Command: "", Term: 0}}
	rs.serverName = serverName
	rs.state = "follower"
	rs.commitIndex = 0
	rs.lastApplied = 0
	rs.clientInput = make(chan string)
	rs.clientOutput = make(chan int)
	rs.rpcLock = &sync.Mutex{}
	fmt.Printf("[SERVER] Server %s is about to register\n", serverName)
	rpc.Register(rs)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("[SERVER][ERROR] Cannot listen on port: ", port)
		fmt.Println(err)
		return nil
	}
	fmt.Printf("[SERVER] Server %s is listening!\n", serverName)
	go http.Serve(l, nil)
	// TODO Maybe sleep here? We need all servers to come online together.
	rs.runStateMachine()
	return rs
}

// Function which follows Raft algorithm according to state.
func (rs *RaftServer) runStateMachine() {
	fmt.Printf("[SERVER] %s about to contact other servers\n", rs.serverName)
	time.Sleep(1000 * time.Millisecond)
	rs.serverMap = make(map[string]*raftRpc.RaftClient)
	for _, s := range rs.cf.Servers {
		if s.Name != "debugServer" && s.Name != rs.serverName {
			addr := s.Address + ":" + strconv.Itoa(s.Port)
			fmt.Printf("[SERVER] %s --> %s\n", rs.serverName, s.Name)
			client, err := raftRpc.DialHTTP("tcp", addr, rs.serverName, s.Name)
			if err != nil {
				fmt.Println("Could not dial ", s.Name)
				fmt.Println(err)
				return
			}
			rs.serverMap[s.Name] = client
		}
	}
	time.Sleep(1000 * time.Millisecond)
	fmt.Printf("[SERVER] %s done contacting other servers\n", rs.serverName)

	for true {
		// If commitIndex > lastApplied:
		// Increment lastApplied, apply log[lastApplied] to state machine.
		for rs.commitIndex > rs.lastApplied {
			// TODO Make persistent
			rs.lastApplied += 1
			rs.pState.StateMachineLog = append(rs.pState.StateMachineLog,
				rs.pState.Log[rs.lastApplied])
			fmt.Printf("[%s] Just applied %s to state machine\n",
				rs.serverName, rs.pState.Log[rs.lastApplied].Command)
		}

		switch rs.state {
		case "follower":
			// Respond to RPCs from candidates + leaders
			// (this is implicit)
			// If the election timeout elapses without receiving
			// AppendEntries RPC from current leader or granting
			// vote to candidate: convert to candidate.
			fmt.Printf("%s : FOLLOWER\n", rs.serverName)
			rs.attemptCandidate = true
			msToWait := 250 + rand.Intn(500)
			time.Sleep(time.Duration(msToWait) * time.Millisecond)
			if rs.attemptCandidate {
				rs.state = "candidate"
			}
		case "candidate":
			fmt.Printf("%s : CANDIDATE\n", rs.serverName)
			// On conversion to candidate, start election.
			// Increment current term
			rs.pState.CurrentTerm += 1
			// Vote for self
			rs.pState.VotedFor = rs.serverName
			numVotes := 1 // We voted for ourselves!
			// Send requestVote RPCs to all other servers
			for s, c := range rs.serverMap {
				var args raftRpc.RaftClientArgs
				var reply raftRpc.RaftClientReply
				args.RequestVoteIn = &raftRpc.RequestVoteInput{}
				args.RequestVoteIn.Term = rs.pState.CurrentTerm
				args.RequestVoteIn.CandidateId = rs.serverName
				args.RequestVoteIn.LastLogIndex = len(rs.pState.Log) - 1
				logIndex := len(rs.pState.Log) - 1
				lastLogEntry := rs.pState.Log[logIndex]
				args.RequestVoteIn.LastLogTerm = lastLogEntry.Term
				if rs.state != "candidate" {
					// If our state became "no longer candidate", break!
					break
				}
				fmt.Printf("[%s][CANDIDATE] Requesting vote from %s\n", rs.serverName, s)
				err := c.Call("RaftServer.RequestVote", args, &reply)
				if err != nil {
					fmt.Println("[%s] Candidate could not request vote", rs.serverName)
				} else {
					if reply.RequestVoteOut == nil {
						fmt.Println("RequestVoteOut was nil")
					} else {
						if reply.RequestVoteOut.VoteGranted {
							fmt.Printf("[SERVER] For %s, from %s, vote granted\n", rs.serverName, s)
							numVotes += 1
						} else if reply.RequestVoteOut.Term > rs.pState.CurrentTerm {
							fmt.Printf("[SERVER] For %s, from %s, vote denied. Requestor is now follower. \n", rs.serverName, s)
							rs.pState.CurrentTerm = reply.RequestVoteOut.Term
							rs.state = "follower"
							break
						}
					}
				}
			}
			numServers := len(rs.serverMap) + 1
			majority := (numServers / 2) + 1
			// If votes received from majority of servers, become leader
			if numVotes >= majority && rs.state == "candidate" {
				fmt.Printf("[%s] Elected Leader!\n", rs.serverName)
				rs.state = "leader"
				// Volatile state must be reinitialized after election.
				rs.nextIndex = make(map[string]int)
				rs.matchIndex = make(map[string]int)
				for s, _ := range rs.serverMap {
					rs.nextIndex[s] = len(rs.pState.Log)
					rs.matchIndex[s] = 0
				}
				if !rs.heartBeat() {
					// Send out initial heartbeat.
					rs.state = "follower"
				}
			} else if rs.state == "candidate" {
				fmt.Printf("[%s] Candidate not elected leader -- backoff.\n", rs.serverName)
				// Backoff for a bit, then try again.
				msToWait := 250 + rand.Intn(500)
				time.Sleep(time.Duration(msToWait) * time.Millisecond)
			}
			// If appendEntries RPC received from new leader, convert
			// to follower
			// (implicit in append entries)
		case "leader":
			fmt.Printf("%s : LEADER\n", rs.serverName)
			// Upon election, send initial empty AppendEntries RPCs
			// (heartbeat) to each server (done), repeat during idle periods
			// to prevent election timeouts.
			select {
			case clientIn := <-rs.clientInput:
				// If command received from client: append entry to local log,
				// respond after entry applied to state machine.
				fmt.Printf("[%s] Leader trying to commit value: %s\n",
					rs.serverName, clientIn)
				entry := &raftRpc.LogEntry{}
				entry.Command = clientIn
				entry.Term = rs.pState.CurrentTerm
				rs.pState.Log = append(rs.pState.Log, entry)
				rs.clientOutput <- len(rs.pState.Log)
			case <-time.After(time.Duration(15) * time.Millisecond):
				fmt.Printf("[%s] Leader outputting heartbeat\n", rs.serverName)
				if !rs.heartBeat() {
					rs.state = "follower"
					break
				}
			}
			// If last log index >= nextIndex for a follower, send
			// AppendEntries RPC with log entries starting at nextIndex.
			for s, _ := range rs.serverMap {
				lastLogIndex := len(rs.pState.Log) - 1
				nextIndex := rs.nextIndex[s]
				if lastLogIndex >= rs.nextIndex[s] {
					rs.sendAppendEntriesRPC(s, rs.pState.Log[nextIndex:])
				}
			}
			// If there exists an N such that N > commitIndex, a majority
			// of matchIndex[i] >= N, and log[N].term == currentTerm,
			// set commitIndex = N.
			n := rs.commitIndex + 1
			numServers := len(rs.serverMap) + 1
			majority := (numServers / 2) + 1
			numReplicated := 0
			if len(rs.pState.Log) > n &&
				(rs.pState.Log[n].Term == rs.pState.CurrentTerm) {
				for s, _ := range rs.serverMap {
					if rs.matchIndex[s] >= n {
						numReplicated += 1
					}
				}
				if numReplicated >= majority {
					fmt.Printf("[%s] Leader has new commitIndex: %d\n",
						rs.serverName, n)
					rs.commitIndex = n
				}
			}
		}
	}
}

/*
Calls the "AppendEntries" RPC for the server S, with the given entries.

returns
 0 on success
-1 on RPC error
-2 on "no longer leader"
*/
func (rs *RaftServer) sendAppendEntriesRPC(s string,
	entries []*raftRpc.LogEntry) int {
	c := rs.serverMap[s]
	fmt.Printf("Sending append entries RPC to %s\n", s)
	var args raftRpc.RaftClientArgs
	var reply raftRpc.RaftClientReply
	args.AppendEntriesIn = &raftRpc.AppendEntriesInput{}
	args.AppendEntriesIn.Term = rs.pState.CurrentTerm
	args.AppendEntriesIn.LeaderId = rs.serverName
	args.AppendEntriesIn.PrevLogIndex = len(rs.pState.Log) - 2
	args.AppendEntriesIn.PrevLogTerm = rs.pState.Log[len(rs.pState.Log)-2].Term
	args.AppendEntriesIn.Entries = entries
	args.AppendEntriesIn.LeaderCommit = rs.commitIndex
	err := c.Call("RaftServer.AppendEntries", args, &reply)
	if err != nil {
		fmt.Println("Leader could not append entries.")
		return -1
	} else if reply.AppendEntriesOut == nil {
		fmt.Println("AppendEntriesOut was nil")
		return -1
	} else if reply.AppendEntriesOut.Term > rs.pState.CurrentTerm {
		rs.pState.CurrentTerm = reply.AppendEntriesOut.Term
		rs.state = "follower"
		return -2
	} else if reply.AppendEntriesOut.Success == false {
		// If AppendEntries fails because of log inconsistency,
		// decrement nextIndex and retry.
		rs.nextIndex[s] -= 1
		nextIndex := rs.nextIndex[s]
		return rs.sendAppendEntriesRPC(s, rs.pState.Log[nextIndex:])
	} else {
		// If successful: Update nextIndex and matchIndex for follower
		rs.nextIndex[s] = args.AppendEntriesIn.PrevLogIndex + len(entries) + 1
		rs.matchIndex[s] = args.AppendEntriesIn.PrevLogIndex + len(entries)
		return 0
	}
}

// Returns "true" on success, "false" if leader
// became follower.
func (rs *RaftServer) heartBeat() bool {
	for s, c := range rs.serverMap {
		fmt.Printf("[%s][HEARTBEAT] to %s\n", rs.serverName, s)
		var args raftRpc.RaftClientArgs
		var reply raftRpc.RaftClientReply
		args.AppendEntriesIn = &raftRpc.AppendEntriesInput{}
		args.AppendEntriesIn.Term = rs.pState.CurrentTerm
		args.AppendEntriesIn.LeaderId = rs.serverName
		args.AppendEntriesIn.PrevLogIndex = len(rs.pState.Log) - 1
		args.AppendEntriesIn.PrevLogTerm = rs.pState.Log[rs.commitIndex].Term
		args.AppendEntriesIn.Entries = nil
		args.AppendEntriesIn.LeaderCommit = rs.commitIndex
		err := c.Call("RaftServer.AppendEntries", args, &reply)
		if err != nil {
			fmt.Println("Leader could not send heartbeat")
		} else {
			if reply.AppendEntriesOut == nil {
				fmt.Println("AppendEntriesOut was nil")
			} else if reply.AppendEntriesOut.Term > rs.pState.CurrentTerm {
				rs.pState.CurrentTerm = reply.AppendEntriesOut.Term
				rs.state = "follower"
				return false
			} // If the heartbeat failed, we can't do anything else.
		}
	}
	return true
}

func (rs *RaftServer) Commit(in string, out *string) error {
	fmt.Printf("[%s] Commit Called\n", rs.serverName)
	rs.clientInput <- in
	entryNum := <-rs.clientOutput
	for len(rs.pState.StateMachineLog) < entryNum {
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("[%s] Responding to client for index %d\n", rs.serverName, entryNum-1)
	*out = rs.pState.StateMachineLog[entryNum-1].Command
	return nil
}

func (rs *RaftServer) ReadLog(index int, value *string) error {
	fmt.Printf("[%s] ReadLog Called, index: %d\n", rs.serverName, index)
	if len(rs.pState.StateMachineLog) > index {
		*value = rs.pState.StateMachineLog[index].Command
		fmt.Println("[RAFT SERVER] ReadLog returning: ", *value)
		return nil
	} else {
		fmt.Printf("[%s] ReadLog error - SML len: %d\n",
			rs.serverName, len(rs.pState.StateMachineLog))
		return errors.New("Index not applied to state machine")
	}
}

func (rs *RaftServer) GetServerStatus(unused int, value *string) error {
	fmt.Println("[RAFT SERVER] GetStatus Called")
	*value = rs.state
	return nil
}

/* Helper function used to call "append entries" on a given server ID */
// TODO MAKE THAT FUNC

func (rs *RaftServer) AppendEntries(in raftRpc.RaftClientArgs,
	out *raftRpc.RaftClientReply) error {
	fmt.Printf("[%s][AE]\n", rs.serverName)
	rs.rpcLock.Lock()
	defer rs.rpcLock.Unlock()
	if in.AppendEntriesIn == nil {
		return errors.New("Need to have AppendEntriesIn argument")
	}
	out.AppendEntriesOut = &raftRpc.AppendEntriesOutput{}

	// 1. Reply false if term < currentTerm
	if in.AppendEntriesIn.Term < rs.pState.CurrentTerm {
		fmt.Printf("[%s][AE] Term %d is less than current %d\n",
			rs.serverName, in.AppendEntriesIn.Term, rs.pState.CurrentTerm)
		out.AppendEntriesOut.Success = false
		out.AppendEntriesOut.Term = rs.pState.CurrentTerm
		return nil
	} else if in.AppendEntriesIn.Term > rs.pState.CurrentTerm {
		fmt.Printf("[%s][AE] Term %d is greater than current %d\n",
			rs.serverName, in.AppendEntriesIn.Term, rs.pState.CurrentTerm)
		rs.pState.CurrentTerm = in.AppendEntriesIn.Term
		rs.state = "follower"
		out.AppendEntriesOut.Term = rs.pState.CurrentTerm
	}
	// 2. Reply false if log doesn't contain an entry at
	// prevLogIndex whose term matches prevLogTerm.
	if (len(rs.pState.Log) < in.AppendEntriesIn.PrevLogIndex) ||
		(rs.pState.Log[in.AppendEntriesIn.PrevLogIndex].Term !=
			in.AppendEntriesIn.PrevLogTerm) {
		fmt.Printf("[%s][AE] Missing prev log entry. Log len: %d\n",
			rs.serverName, len(rs.pState.Log))
		out.AppendEntriesOut.Success = false
		return nil
	}
	// 3. If an existing entry conflicts with a new one (same
	// index but different terms), delete the existing entry and all
	// that follow it.
	startIndex := in.AppendEntriesIn.PrevLogIndex + 1
	for i := startIndex; i < startIndex+len(in.AppendEntriesIn.Entries); i++ {
		fmt.Printf("[%s][AE] Checking conflicts at index %d\n", rs.serverName, i)
		if len(rs.pState.Log)-1 < i {
			fmt.Printf("[%s][AE] End of conflicts at index %d\n", rs.serverName, i)
			break
		} else if rs.pState.Log[i].Term != in.AppendEntriesIn.Term {
			fmt.Printf("[%s][AE] Cutting log size to: %d\n", rs.serverName, i)
			// TODO make persistent?
			rs.pState.Log = rs.pState.Log[:i]
			break
		}
	}
	// 4. Append any new entries not already in the log.
	for i := startIndex; i < startIndex+len(in.AppendEntriesIn.Entries); i++ {
		fmt.Printf("[%s][AE] Adding entry at index %d\n", rs.serverName, i)
		// TODO make persistent?
		rs.pState.Log = append(rs.pState.Log, in.AppendEntriesIn.Entries[i-startIndex])
		for _, e := range rs.pState.Log {
			fmt.Println(e.Command)
		}
	}
	// 5. If leaderCommit > commitIndex, set
	// commitIndex = min(leaderCommit, index of last new entry)
	lastIndex := startIndex + len(in.AppendEntriesIn.Entries) - 1
	if in.AppendEntriesIn.LeaderCommit > rs.commitIndex {
		fmt.Printf("[%s][AE] Updating leadercommit. LC: %d. LI: %d.\n",
			rs.serverName, in.AppendEntriesIn.LeaderCommit, lastIndex)
		if in.AppendEntriesIn.LeaderCommit <= lastIndex {
			rs.commitIndex = in.AppendEntriesIn.LeaderCommit
		} else {
			rs.commitIndex = lastIndex
		}
		fmt.Printf("[%s][AE] Commit index: %d \n", rs.serverName, rs.commitIndex)
	}
	rs.attemptCandidate = false
	out.AppendEntriesOut.Success = true
	return nil
}

func (rs *RaftServer) RequestVote(in raftRpc.RaftClientArgs,
	out *raftRpc.RaftClientReply) error {
	rs.rpcLock.Lock()
	defer rs.rpcLock.Unlock()
	fmt.Printf("[%s][RV] LOCK ACQUIRED\n", rs.serverName)
	if in.RequestVoteIn == nil {
		return errors.New("Need to have requestVoteIn argument")
	}
	out.RequestVoteOut = &raftRpc.RequestVoteOutput{}
	out.RequestVoteOut.VoteGranted = false

	fmt.Printf("[%s][RV][CurrentTerm: %d][RequestTerm: %d]\n",
		rs.serverName, rs.pState.CurrentTerm, in.RequestVoteIn.Term)
	// 1. Reply false if term < currentTerm
	if in.RequestVoteIn.Term < rs.pState.CurrentTerm {
		fmt.Printf("[%s][RV] Term too small\n", rs.serverName)
		out.RequestVoteOut.Term = rs.pState.CurrentTerm
		return nil
	} else if in.RequestVoteIn.Term > rs.pState.CurrentTerm {
		fmt.Printf("[%s][RV] This term is large!\n", rs.serverName)
		rs.pState.CurrentTerm = in.RequestVoteIn.Term
		rs.pState.VotedFor = ""
		rs.state = "follower"
		out.RequestVoteOut.Term = rs.pState.CurrentTerm
	}
	// 2. If votedFor is null or candidateId, and candidate's log
	// is at least as up-to-date as receiver's log, grant vote.
	fmt.Printf("[%s][RV][VotedFor: %s][Loglen: %d][LLI: %d]\n",
		rs.serverName, rs.pState.VotedFor, len(rs.pState.Log), in.RequestVoteIn.LastLogIndex)
	if ((rs.pState.VotedFor == "") ||
		(rs.pState.VotedFor == in.RequestVoteIn.CandidateId)) &&
		(len(rs.pState.Log)-1 <= in.RequestVoteIn.LastLogIndex) {
		fmt.Printf("[%s][RV] Vote granted to %s\n", rs.serverName, in.RequestVoteIn.CandidateId)
		rs.pState.VotedFor = in.RequestVoteIn.CandidateId
		out.RequestVoteOut.VoteGranted = true
		rs.attemptCandidate = false
	}
	return nil
}

func GimmeTrue() bool {
	return true
}
