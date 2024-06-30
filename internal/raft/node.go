package raft

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync/atomic"

	"sync"
	"time"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/models"
	"github.com/Deathfireofdoom/distributed-kv-store/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RaftNode struct {
	proto.UnimplementedRaftServiceServer
	mu          sync.Mutex
	id          string
	term        int32
	votedFor    string
	log         []LogEntry
	commitIndex int32
	lastApplied int32
	nextIndex   map[string]int32
	matchIndex  map[string]int32
	fsm         StateMachine
	peers       []string
	isLeader    bool
	votes       int32
	heartbeatCh chan bool
}

func NewRaftNode(id string, peers []string, fsm StateMachine) *RaftNode {
	node := &RaftNode{
		id:          id,
		log:         make([]LogEntry, 0),
		nextIndex:   make(map[string]int32),
		matchIndex:  make(map[string]int32),
		fsm:         fsm,
		peers:       peers,
		heartbeatCh: make(chan bool),
	}
	go node.startElectionTimer()
	return node
}

func (node *RaftNode) StartGRPCServer(port string) error {
	server := grpc.NewServer()
	proto.RegisterRaftServiceServer(server, node)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	return server.Serve(lis)
}

// local method - Invoked by client sending the http request
func (node *RaftNode) PutHanlder(w http.ResponseWriter, r *http.Request) {
	var req models.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	command := Command{
		Type:  PutCommand,
		Key:   req.Key,
		Value: req.Value,
	}

	data, _ := json.Marshal(command)
	entry := LogEntry{
		Term:    node.term,
		Command: string(data),
	}

	node.mu.Lock()
	node.log = append(node.log, entry)
	node.mu.Unlock()

	// asking the rest of the nodes to add the entry to their log
	// but not commit it, so not updating their kv-store
	success := node.replicateLogEntry(entry)
	if success {
		// applying the changes to local kv-store
		node.mu.Lock()
		node.commitIndex = int32(len(node.log) - 1)
		node.applyLogEntries()
		node.mu.Unlock()

		// asking the nodes to commit, update their kv store with the logs
		go node.notifyCommit()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.PutResponse{Success: true})
	} else {
		http.Error(w, "failed to replicate log entry", http.StatusInternalServerError)
	}
}

func (node *RaftNode) replicateLogEntry(entry LogEntry) bool {
	panic("implement")
}

func (node *RaftNode) notifyCommit() {
	panic("implement")
}

// RAFT METHODS - HEART BEATS
func (node *RaftNode) startHeartbeat() {
	log.Printf("%s sending heartbeats, and are leader: %t", node.id, node.isLeader)
	for node.isLeader {
		node.sendHeartBeats()
		time.Sleep(5 * time.Second)
	}
}

func (node *RaftNode) sendHeartBeats() {
	for _, peer := range node.peers {
		go node.sendHeartbeatToPeer(peer)
	}
}

func (node *RaftNode) sendHeartbeatToPeer(peer string) {
	// debug logging
	log.Printf("%s sending heart beet to %s", node.id, peer)

	// A heartbeat is simply a appendRequest without any new logs
	conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials())) // figure out timeout
	if err != nil {
		return
	}
	defer conn.Close()

	client := proto.NewRaftServiceClient(conn)

	var lastLogIndex int32
	var lastLogTerm int32
	if len(node.log) > 0 {
		lastLogIndex = int32(len(node.log) - 1)
		lastLogTerm = node.log[lastLogIndex].Term
	}

	req := &proto.AppendEntriesRequest{
		Term:         node.term,
		LeaderId:     node.id,
		PrevLogIndex: lastLogIndex,
		PrevLogTerm:  lastLogTerm,
		LeaderCommit: node.commitIndex,
	}

	client.AppendEntries(context.Background(), req)
}

// RAFT METHODS - Leader election
func (node *RaftNode) startElectionTimer() {
	// this functions has a timer that will invoke leader election.
	// Everytime the node gets a heartbeat from the leader the timer
	// resets.
	for {
		// the timeout has a base-timeout, and a random addition so
		// not all nodes timeout at the same time causing a stampeede.
		timeout := 10*time.Second + time.Duration(rand.Intn(30))*time.Second
		if node.id == "node1" {
			timeout = 1 * time.Second
		}
		timer := time.NewTimer(timeout)

		select {
		case <-timer.C:
			node.mu.Lock()
			if !node.isLeader {
				node.startElection()
			}
			node.mu.Unlock()
		case <-node.heartbeatCh:
			timer.Stop()
		}
	}

}

func (node *RaftNode) startElection() {
	node.term++
	node.votedFor = node.id
	node.votes = 1 // this also resets
	node.isLeader = false

	for _, peer := range node.peers {
		go node.requestVoteFromPeer(peer)
	}

	// wait for 5 second before counting votes
	time.Sleep(2 * time.Second)
	if node.votes > int32(len(node.peers))/2 {
		log.Printf("%s became the leader", node.id)
		node.isLeader = true
		go node.startHeartbeat()
	}
	log.Printf("%s got %d votes", node.id, node.votes)
}

func (node *RaftNode) requestVoteFromPeer(peer string) {
	// debug print
	log.Printf("%s is request vote from %s", node.id, peer)

	conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials())) // figure out timeout
	if err != nil {
		return
	}
	defer conn.Close()

	client := proto.NewRaftServiceClient(conn)

	var lastLogIndex int32
	var lastLogTerm int32
	if len(node.log) > 0 {
		lastLogIndex = int32(len(node.log) - 1)
		lastLogTerm = node.log[lastLogIndex].Term
	}

	req := &proto.RequestVoteRequest{
		Term:         node.term,
		CandidateId:  node.id,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}

	resp, err := client.RequestVote(context.Background(), req)
	if err != nil {
		log.Printf("failed to request vote from peer %s: %v", peer, err)
	}

	log.Printf("%s got this response from peer %s: %v", node.id, peer, resp)
	if resp.VoteGranted {
		atomic.AddInt32(&node.votes, 1)
	}

	if resp.Term > node.term {
		node.term = resp.Term
		node.votedFor = ""
		node.isLeader = false
	}
}

// gRPC Method - invoked by other node
func (node *RaftNode) AppendEntries(ctx context.Context, req *proto.AppendEntriesRequest) (*proto.AppendEntriesResponse, error) {
	log.Printf("%s got a entry: %v", node.id, req)
	node.mu.Lock()
	defer node.mu.Unlock()

	// Purpose: Check if the incoming request is comming from a outdated "leader"
	// Desc: 	If the node.term is greater than the term of the sender, then it means
	//			the "sender" is outdated, so can't accept the log entry.
	if req.Term < node.term {
		return &proto.AppendEntriesResponse{
			Term:    node.term,
			Success: false,
		}, nil
	}

	// Purpose: Recognize the authority of the sender.
	// Desc: 	Updating the current node to reflect that
	//			it accepts the sender as the leader.
	node.term = req.Term
	node.votedFor = req.LeaderId
	node.isLeader = false

	// Purpose: Reseting timer for leader election
	// Desc:	This means we got a heartbeat from the leader,
	//			so we need to reset the heartbeat timer.
	select {
	case node.heartbeatCh <- true:
	default:
	}

	// Purpose: Check for consitency in logs
	// Desc:	If the prevLog is not populated in the node
	//			or it has another term number, then something
	//			is off and can't accept the entry.
	if len(node.log) < int(req.PrevLogIndex)+1 || node.log[req.PrevLogIndex].Term != req.PrevLogTerm {
		return &proto.AppendEntriesResponse{
			Term:    node.term,
			Success: false,
		}, nil
	}

	// Purpose: Add the new logs to the node log
	// Desc:	Since we accepted the logs we also need to add them to
	//			our log to make sure consitency
	node.log = append(node.log[:req.PrevLogIndex+1], convertProtoEntries(req.Entries)...)

	// Purpose: Commits logs that has been commited by the leader
	// Desc:	This is a bit confusing, but the important thing to remember
	//			is that the "append log"-part and this commit part will most likely
	//			not be ran in the same execution. The leader first sends the logs with
	//			a non-updated commit index, then it sends a another request, without logs,
	//			but just a updated commit index.
	if req.LeaderCommit > node.commitIndex {
		node.commitIndex = min(req.LeaderCommit, int32(len(node.log)+1))
		node.applyLogEntries()
	}

	// Purpose: The log as been accepted, so we retutn that to the leader
	return &proto.AppendEntriesResponse{
		Term:    node.term,
		Success: true,
	}, nil
}

// gRPC Method - invoked by other node
func (node *RaftNode) RequestVote(ctx context.Context, req *proto.RequestVoteRequest) (*proto.RequestVoteResponse, error) {
	// debug print
	log.Printf("%s has request %s to vote", req.CandidateId, node.id)

	node.mu.Lock()
	defer node.mu.Unlock()
	// This function is invoked when another node ask to become the leader.

	// Purpose: To see if the node asking to become leader is less updated
	// Desc:	If node.term is higher than req.Term, then it means the node
	//			asking to become the leader is not as updated as the current
	//			node, so the "requester" cant become the leader.
	if req.Term < node.term {
		log.Printf("%s does not accept %s as leader due to term %d vs %d", node.id, req.CandidateId, req.Term, node.term)
		return &proto.RequestVoteResponse{
			Term:        node.term,
			VoteGranted: false,
		}, nil
	}

	if req.Term > node.term {
		node.term = req.Term
		node.votedFor = ""
	}

	// Purpose: Check if the node already voted on someone else in the same term
	// Desc:	Several nodes can ask to become leader, but a node is only allowed
	//			to vote once in a term.
	if node.votedFor == "" || node.votedFor == req.CandidateId {
		log.Printf("%s accept %s as leader", node.id, req.CandidateId)
		node.term = req.Term
		node.votedFor = req.CandidateId
		return &proto.RequestVoteResponse{
			Term:        node.term,
			VoteGranted: true,
		}, nil
	}

	// Purpose: Not accepting the vote, this happens if we already voted
	log.Printf("%s does not accept %s as leader already voted for %s", node.id, req.CandidateId, node.votedFor)
	return &proto.RequestVoteResponse{
		Term:        node.term,
		VoteGranted: false,
	}, nil
}

func (node *RaftNode) applyLogEntries() {
	// actually applying the logs to the state machine,
	// in this case it means updating the kv-store.
	// This is what we call "commiting".
	for node.lastApplied < node.commitIndex {
		node.lastApplied++
		entry := node.log[node.lastApplied]
		var command Command
		json.Unmarshal([]byte(entry.Command), &command)
		node.fsm.Apply(command)
	}
}

func convertProtoEntries(entries []*proto.LogEntry) []LogEntry {
	logEntries := make([]LogEntry, len(entries))
	for i, entry := range entries {
		logEntries[i] = LogEntry{
			Term:    entry.Term,
			Command: entry.Command,
		}
	}
	return logEntries
}

func convertToProtoEntries(entries []LogEntry) []*proto.LogEntry {
	protoEntries := make([]*proto.LogEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = &proto.LogEntry{
			Term:    entry.Term,
			Command: entry.Command,
		}
	}
	return protoEntries
}
