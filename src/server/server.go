package server

import (
	"config"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"raftPersistency"
	"raftRpc"
	"strconv"
)

/* All state on all servers */
type RaftServer struct {
	serverName string
	cf         config.Config
	serverMap  map[int]*raftRpc.RaftClient

	pState      *raftPersistency.RaftPersistentState
	state       string /* "follower", "candidate", "leader" */
	commitIndex int
	lastApplied int
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
	rs.pState.VotedFor = -1
	rs.serverName = serverName
	rs.state = "follower"
	rpc.Register(rs)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil
	}
	http.Serve(l, nil)
	go rs.RunStateMachine()
	return rs
}

// Function which follows Raft algorithm according to state.
func (rs *RaftServer) RunStateMachine() {
	for true {

	}
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
	if ((rs.pState.VotedFor == -1) ||
		(rs.pState.VotedFor == in.RequestVoteIn.CandidateId)) &&
		(len(rs.pState.Log) <= in.RequestVoteIn.LastLogIndex) {
		out.RequestVoteOut.VoteGranted = true
	}
	return nil
}

func GimmeTrue() bool {
	return true
}
