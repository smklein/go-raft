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
	"time"
)

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
	nextIndex  []int
	matchIndex []int
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
	rs := &RaftServer{}
	if !config.LoadConfig(&rs.cf) {
		return nil
	}
	port := 0
	for _, s := range rs.cf.Servers {
		if s.Name == serverName {
			port = s.Port
		}
	}
	if port == 0 {
		return nil
	}
	// TODO Store/Load!
	rs.pState = &raftPersistency.RaftPersistentState{}
	rs.pState.VotedFor = ""
	rs.serverName = serverName
	rs.state = "follower"
	rpc.Register(rs)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil
	}
	http.Serve(l, nil)
	// TODO Maybe sleep here? We need all servers to come online together.
	rs.serverMap = make(map[string]*raftRpc.RaftClient)
	for _, s := range rs.cf.Servers {
		if s.Name != "debugServer" && s.Name != serverName {
			addr := s.Address + ":" + strconv.Itoa(s.Port)
			client, err := raftRpc.DialHTTP("tcp", addr)
			if err != nil {
				fmt.Println("Could not dial ", s.Name)
				return nil
			}
			rs.serverMap[s.Name] = client
		}
	}

	go rs.RunStateMachine()
	return rs
}

// Function which follows Raft algorithm according to state.
func (rs *RaftServer) RunStateMachine() {
	for true {
		// If commitIndex > lastApplied:
		// Increment lastApplied, apply log[lastApplied] to state machine.
		if rs.commitIndex > rs.lastApplied {
			// TODO Make persistent
			rs.pState.StateMachineLog[rs.lastApplied] =
				rs.pState.Log[rs.lastApplied]
			rs.lastApplied += 1
		}

		switch rs.state {
		case "follower":
			// Respond to RPCs from candidates + leaders
			// (this is implicit)
			// If the election timeout elapses without receiving
			// AppendEntries RPC from current leader or granting
			// vote to candidate: convert to candidate.
			rs.attemptCandidate = true
			msToWait := 25 + rand.Intn(25)
			<-time.After(time.Duration(msToWait) * time.Millisecond)
			if rs.attemptCandidate {
				rs.state = "candidate"
			}
		case "candidate":
			// On conversion to candidate, start election.
			// Increment current term
			rs.pState.CurrentTerm += 1
			// Vote for self
			rs.pState.VotedFor = rs.serverName
			// Reset election timer
			// TODO
			numVotes := 1 // We voted for ourselves!
			// Send requestVote RPCs to all other servers
			for s, c := range rs.serverMap {
				fmt.Println("Sending request vote RPC to ", s)
				var args raftRpc.RaftClientArgs
				var reply raftRpc.RaftClientReply
				args.RequestVoteIn = &raftRpc.RequestVoteInput{}
				args.RequestVoteIn.Term = rs.pState.CurrentTerm
				args.RequestVoteIn.CandidateId = rs.serverName
				args.RequestVoteIn.LastLogIndex = len(rs.pState.Log) - 1
				if len(rs.pState.Log) > 0 {
					logIndex := len(rs.pState.Log) - 1
					lastLogEntry := rs.pState.Log[logIndex]
					args.RequestVoteIn.LastLogTerm = lastLogEntry.Term
				} else {
					args.RequestVoteIn.LastLogTerm = -1
				}
				if rs.state != "candidate" {
					// If our state became "no longer candidate", break!
					break
				}
				err := c.Call("RaftServer.RequestVote", args, &reply)
				if err != nil {
					fmt.Println("Candidate could not request vote")
				} else {
					if reply.RequestVoteOut == nil {
						fmt.Println("RequestVoteOut was nil")
					} else {
						if reply.RequestVoteOut.VoteGranted {
							numVotes += 1
						} else if reply.RequestVoteOut.Term > rs.pState.CurrentTerm {
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
				rs.state = "leader"
			}
			// If appendEntries RPC received from new leader, convert
			// to follower
			// (implicit in append entries)
		case "leader":
			// TODO
			// TODO Heartbeats
			// If command received from client: append entry to local log,
			// respond after entry applied to state machine.
			// TODO
			// If last log index >= nextIndex for a follower, send
			// AppendEntries RPC with log entries starting at nextIndex.
			// TODO
			// If successful: Update nextIndex and matchIndex for follower
			// If AppendEntries fails because of log inconsistency,
			// decrement nextIndex and retry.
			// TODO
			// If there exists an N such that N > commitIndex, a majority
			// of matchIndex[i] >= N, and log[N].term == currentTerm,
			// set commitIndex = N.
			return
		}
	}
}

// TODO
func (rs *RaftServer) HeartBeat() {

}

// TODO Add ability to call this function in client.
// TODO Add other RPCs needed by client.
func (rs *RaftServer) Commit(in string,
	out *string) error {
	fmt.Println("[RAFT SERVER] Commit Called")
	return nil
}

/* Helper function used to call "append entries" on a given server ID */
// TODO MAKE THAT FUNC

func (rs *RaftServer) AppendEntries(in raftRpc.RaftClientArgs,
	out *raftRpc.RaftClientReply) error {
	if in.AppendEntriesIn == nil {
		return errors.New("Need to have AppendEntriesIn argument")
	}
	out.AppendEntriesOut = &raftRpc.AppendEntriesOutput{}

	// 1. Reply false if term < currentTerm
	if in.AppendEntriesIn.Term < rs.pState.CurrentTerm {
		out.AppendEntriesOut.Success = false
		out.AppendEntriesOut.Term = rs.pState.CurrentTerm
		return nil
	} else if in.AppendEntriesIn.Term > rs.pState.CurrentTerm {
		rs.pState.CurrentTerm = in.AppendEntriesIn.Term
		rs.state = "follower"
		out.AppendEntriesOut.Term = rs.pState.CurrentTerm
	}
	// 2. Reply false if log doesn't contain an entry at
	// prevLogIndex whose term matches prevLogTerm.
	if (len(rs.pState.Log) < in.AppendEntriesIn.PrevLogIndex) ||
		(rs.pState.Log[in.AppendEntriesIn.PrevLogIndex].Term !=
			in.AppendEntriesIn.PrevLogTerm) {
		out.AppendEntriesOut.Success = false
		return nil
	}
	// 3. If an existing entry conflicts with a new one (same
	// index but different terms), delete the existing entry and all
	// that follow it.
	startIndex := in.AppendEntriesIn.PrevLogIndex + 1
	for i := startIndex; i < startIndex+len(in.AppendEntriesIn.Entries); i++ {
		if len(rs.pState.Log) < i {
			break
		} else if rs.pState.Log[i].Term != in.AppendEntriesIn.Term {
			// TODO make persistent?
			rs.pState.Log = rs.pState.Log[:i]
			break
		}
	}
	// 4. Append any new entries not already in the log.
	for i := startIndex; i < len(in.AppendEntriesIn.Entries); i++ {
		// TODO make persistent?
		rs.pState.Log[i] = in.AppendEntriesIn.Entries[i-startIndex]
	}
	// 5. If leaderCommit > commitIndex, set
	// commitIndex = min(leaderCommit, index of last new entry)
	lastIndex := startIndex + len(in.AppendEntriesIn.Entries) - 1
	if in.AppendEntriesIn.LeaderCommit > rs.commitIndex {
		if in.AppendEntriesIn.LeaderCommit <= lastIndex {
			rs.commitIndex = in.AppendEntriesIn.LeaderCommit
		} else {
			rs.commitIndex = lastIndex
		}
	}
	rs.attemptCandidate = false
	out.AppendEntriesOut.Success = true
	return nil
}

func (rs *RaftServer) RequestVote(in raftRpc.RaftClientArgs,
	out *raftRpc.RaftClientReply) error {
	if in.RequestVoteIn == nil {
		return errors.New("Need to have requestVoteIn argument")
	}
	out.RequestVoteOut = &raftRpc.RequestVoteOutput{}
	out.RequestVoteOut.VoteGranted = false

	// 1. Reply false if term < currentTerm
	if in.RequestVoteIn.Term < rs.pState.CurrentTerm {
		out.RequestVoteOut.Term = rs.pState.CurrentTerm
		return nil
	} else if in.RequestVoteIn.Term > rs.pState.CurrentTerm {
		rs.pState.CurrentTerm = in.RequestVoteIn.Term
		rs.state = "follower"
		out.RequestVoteOut.Term = rs.pState.CurrentTerm
	}
	// 2. If votedFor is null or candidateId, and candidate's log
	// is at least as up-to-date as receiver's log, grant vote.
	if ((rs.pState.VotedFor == "") ||
		(rs.pState.VotedFor == in.RequestVoteIn.CandidateId)) &&
		(len(rs.pState.Log) <= in.RequestVoteIn.LastLogIndex) {
		out.RequestVoteOut.VoteGranted = true
		rs.attemptCandidate = false
	}
	return nil
}

func GimmeTrue() bool {
	return true
}
