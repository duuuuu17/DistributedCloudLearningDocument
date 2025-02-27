package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"

	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labgob"
	"6.5840/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 3D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 3D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (3A, 3B, 3C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	whoami NodeType
	// election
	Stat             State
	lastElectionTime int64
	electionTimeout  int64
	// log
	applyCh   chan ApplyMsg
	applyCond *sync.Cond
	commit    chan struct{}
}
type NodeType int8

const (
	Leader NodeType = iota
	Candidate
	Follower
)

type State struct {
	currentTerm    int
	votedFor       int
	log            []LogEntries
	committedIndex int // 已提交日志项的索引
	lastApplied    int // 已应用到状态机的最高日志项索引
	leaderState    LeaderState
}
type LeaderState struct {
	nextIndex  []int // 记录着不同追随者下一个发送的日志条目的索引
	matchIndex []int // 记录着不同追随者已经成功复制(发送)的日志条目的索引
}
type LogEntries struct {
	Term    int
	Command interface{}
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// Your code here (3A).
	term = rf.Stat.currentTerm
	isleader = rf.whoami == Leader
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	buf := new(bytes.Buffer)
	encoder := labgob.NewEncoder(buf)
	encoder.Encode(rf.Stat.currentTerm)
	encoder.Encode(rf.Stat.votedFor)
	encoder.Encode(rf.Stat.log)
	nodeStat := buf.Bytes()
	rf.persister.Save(nodeStat, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	var log []LogEntries
	var term, votedFor int
	readBuf := bytes.NewBuffer(data)
	decoder := labgob.NewDecoder(readBuf)
	if decoder.Decode(&term) != nil ||
		decoder.Decode(&votedFor) != nil ||
		decoder.Decode(&log) != nil {
		DPrintf("Persistantly data is nil!")
	}
	// DPrintf("Read data, node:%v term:%v vaotefor:%v log:%v!", node, term, votedFor, log)
	rf.Stat.currentTerm = term
	rf.Stat.votedFor = votedFor
	rf.Stat.log = log

}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// 候选者发起选举投票
func (rf *Raft) CandidateRequestVote() {

	lastLogIndex := rf.getLastLogIndex()
	args := new(RequestVoteArgs)
	args.Term, args.CandidateId,
		args.LastLogIndex, args.LastLogTerm =
		rf.Stat.currentTerm, rf.me, lastLogIndex, rf.getLogTerm(lastLogIndex)

	voteCounts := 1
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		go func(i int) {
			reply := new(RequestVoteReply)
			if ok := rf.sendRequestVote(i, args, reply); ok {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				if args.Term == rf.Stat.currentTerm && rf.whoami == Candidate {
					// DPrintf("{Node %v} receives RequestVoteReply %v from {Node %v} after sending RequestVoteArgs %v", rf.me, reply, i, args)
					if reply.Term > rf.Stat.currentTerm {
						rf.convertToFollower(reply.Term)
						return
					}
					// 当回复成功
					if reply.VoteGranted {
						voteCounts++
						// 并且投票超过半数和当前raft节点为候选者时
						if voteCounts > (len(rf.peers) / 2) {
							// DPrintf("the node:%v turn leader!", rf.me)
							rf.convertToLeader()
						}
						// 当候选者得到的跟随者回复的任期比他更大时
					}
				}
			}
		}(i)
	}
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)

	return ok

}

// 处理返回的AppendEntries请求
func (rf *Raft) processAppendEntriesReply(server int, isHeartbeat bool) {
	rf.mu.Lock()
	nextIndex := rf.Stat.leaderState.nextIndex[server]
	prevLogIndex := nextIndex - 1
	prevLogTerm := rf.getLogTerm(prevLogIndex)
	entries := make([]LogEntries, 0)
	if nextIndex < len(rf.Stat.log) {
		entries = append(entries, rf.Stat.log[nextIndex:]...)
	} else if isHeartbeat {
		entries = nil
	}
	args := &AppendEntriesArgs{
		rf.Stat.currentTerm,
		rf.me,
		prevLogIndex,
		prevLogTerm,
		entries,
		rf.Stat.committedIndex,
	}
	reply := &AppendEntriesReply{}
	rf.mu.Unlock()
	if ok := rf.sendAppendEntries(server, args, reply); ok {
		rf.mu.Lock()
		defer rf.mu.Unlock()
		// 忽略过期任期的响应
		if args.Term != rf.Stat.currentTerm || rf.whoami != Leader {
			if reply.Term > rf.Stat.currentTerm {
				DPrintf("turn follower")
				// 跟随者的任期比自己大，自己是旧leader。需转变为follower
				rf.convertToFollower(reply.Term)
				return
			}
			return
		}
		if !reply.Success {
			// 处理冲突任期
			if reply.ConflictTerm == -1 {
				rf.Stat.leaderState.nextIndex[server] = reply.ConflictIndex
			} else { // 否则表示leader的nextIndex记录的过大，需要将往早期日志项移动
				lastConflictIndexInTerm := -1
				for i := len(rf.Stat.log) - 1; i >= 0; i-- {
					if reply.ConflictTerm == rf.Stat.log[i].Term {
						lastConflictIndexInTerm = i
						break
					}
				}
				//  若存在冲突任期的日志项，且根据Raft协议规则，表明该冲突任期的所有日志项是被保存到多数服务器的
				// 此时只需要覆盖该冲突任期之后的日志项
				// 		10	11	12	13
				// S1 	3	4	4	[ ]		Leader: S2 Follower: S1
				// S2	3	4	5	[6]		prevLogIndex: 12 preLogTerm: 5
				// After conflicted handle:
				// 		10	11	12	13		conflictIndex: 11, conflictTerm: 4, nextIndex: 12
				// S1 	3	4	[4]			Leader: S2 Follower: S1
				// S2	3	4	[5	6]		prevLogIndex: 11 preLogTerm: 4
				if lastConflictIndexInTerm != -1 {
					rf.Stat.leaderState.nextIndex[server] = lastConflictIndexInTerm + 1
				} else { // 否则，表明follower的该冲突任期的日志项需要被leader的日志项进行覆盖
					rf.Stat.leaderState.nextIndex[server] = reply.ConflictIndex
				}
			}
		} else { //当成功将log复制到当前遍历的raft节点
			if args.Entries != nil {
				rf.Stat.leaderState.matchIndex[server] = args.PrevLogIndex + len(args.Entries)
				rf.Stat.leaderState.nextIndex[server] = rf.Stat.leaderState.matchIndex[server] + 1
				// DPrintf("Leader:%v	 Exsist logs:%v", rf.me, rf.Stat.log)
				rf.checkAndCommitLogs() // 检查复制日志项，并准备提交
			}
		}
	}
}

// 在Client向领导者请求命令时，需要遍历所有follower复制命令
// 复制日志给follower
func (rf *Raft) replicateLogsToFollower(isHeartbeat bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	for server := range rf.peers {
		if server == rf.me {
			continue
		}
		go rf.processAppendEntriesReply(server, isHeartbeat)

	}
}
func (rf *Raft) heartbeat() {
	for !rf.killed() {
		rf.mu.Lock()
		if rf.whoami != Leader {
			rf.mu.Unlock()
			return
		}
		go rf.replicateLogsToFollower(true)
		rf.mu.Unlock()
		time.Sleep(60 * time.Millisecond)
	}

}

func (rf *Raft) checkAndCommitLogs() {

	majority := len(rf.peers)/2 + 1
	// for n := rf.Stat.committedIndex + 1; n <= rf.getLastLogIndex(); n++ {
	for n := rf.getLastLogIndex(); n > rf.Stat.committedIndex; n-- { // 按raft规则来说，每当提交时，会默认之前任期已经提交完成，所以倒序索引
		if rf.Stat.log[n].Term != rf.Stat.currentTerm {
			continue
		}
		count := 1 // leader vote
		for peer := range rf.peers {
			if peer != rf.me && rf.Stat.leaderState.matchIndex[peer] >= n {
				count++
			}
		}
		if count >= majority {
			rf.Stat.committedIndex = n
			// DPrintf("Leader Broadcast")
			rf.applyCond.Broadcast() // 将当前提交应用到状态机
			break
		}
	}
}

func (rf *Raft) applyCommittedLogs() {
	for !rf.killed() {
		rf.mu.Lock()
		for rf.Stat.lastApplied >= rf.Stat.committedIndex {
			rf.applyCond.Wait()
		}
		// DPrintf("lastApplied:%v, CommittedIndex:%v", rf.Stat.lastApplied, rf.Stat.committedIndex)
		applyStart := rf.Stat.lastApplied + 1
		applyEnd := rf.Stat.committedIndex
		entries := make([]LogEntries, applyEnd-applyStart+1)
		copy(entries, rf.Stat.log[applyStart:applyEnd+1])
		rf.mu.Unlock()

		for i, entry := range entries {
			// 这里可以实现将日志条目应用到状态机的逻辑
			// 例如，如果是键值对存储，可以执行相应的读写操作
			// 处理 command
			rf.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      entry.Command,
				CommandIndex: applyStart + i,
			}
		}
		rf.mu.Lock()
		if applyEnd > rf.Stat.lastApplied {
			rf.Stat.lastApplied = applyEnd
		}
		rf.mu.Unlock()
	}

	// DP.Printf("this apply the entries index:%v\n", rf.Stat.lastApplied)
}
func (rf *Raft) convertToFollower(term int) {

	rf.Stat.currentTerm = term
	rf.whoami = Follower
	rf.Stat.votedFor = -1

	rf.persist()
	go rf.resetElectionTimer()
}
func (rf *Raft) convertToLeader() {
	rf.whoami = Leader
	for i := range rf.peers {
		rf.Stat.leaderState.nextIndex[i] = len(rf.Stat.log) + 1
		rf.Stat.leaderState.matchIndex[i] = 0
	}
	go rf.heartbeat()

}
func (rf *Raft) getLastLogIndex() int {
	if len(rf.Stat.log) == 0 {
		return 0
	}
	return len(rf.Stat.log) - 1
}
func (rf *Raft) getLogTerm(index int) int {
	if index == -1 || index < 0 || index >= len(rf.Stat.log) {
		return 0
	}
	return rf.Stat.log[index].Term
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := -1
	isLeader := rf.whoami == Leader
	// Your code here (3B).
	if !isLeader {
		return index, term, isLeader
	}
	term = rf.Stat.currentTerm
	newEntry := LogEntries{
		Term:    rf.Stat.currentTerm,
		Command: command,
	}

	rf.Stat.log = append(rf.Stat.log, newEntry)
	rf.persist()

	index = rf.getLastLogIndex()
	// DPrintf("concurrent start nodeID:%v", rf.me)
	return index, term, isLeader
}
func (rf *Raft) StartElection() {

	rf.Stat.votedFor = rf.me
	rf.whoami = Candidate
	rf.Stat.currentTerm++
	rf.persist()
	rf.resetElectionTimer()
	rf.CandidateRequestVote()

}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for !rf.killed() {

		// Your code here (3A)
		// Check if a leader election should be started.
		rf.mu.Lock()
		if rf.whoami != Leader && rf.isElectionTimeout() {
			rf.StartElection()
		}
		rf.mu.Unlock()
		ms := 100
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}
func getRandResetElectionTime() time.Duration {
	return time.Duration((300 + rand.Int63()%150)) * time.Millisecond
}
func (rf *Raft) isElectionTimeout() bool {
	return time.Now().UnixMilli()-atomic.LoadInt64(&rf.lastElectionTime) > atomic.LoadInt64(&rf.electionTimeout)
}
func (rf *Raft) resetElectionTimer() {
	atomic.StoreInt64(&rf.lastElectionTime, time.Now().UnixMilli())
	atomic.StoreInt64(&rf.electionTimeout, getRandResetElectionTime().Milliseconds())
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3A, 3B, 3C).

	rf.whoami = Follower
	rf.applyCh = applyCh
	rf.applyCond = sync.NewCond(&rf.mu)
	rf.commit = make(chan struct{}, 1)
	rf.Stat = State{
		currentTerm:    0,
		votedFor:       -1,
		log:            make([]LogEntries, 1),
		committedIndex: 0,
		lastApplied:    0,
		leaderState: LeaderState{
			make([]int, len(peers)),
			make([]int, len(peers)),
		},
	}
	rf.Stat.log[0] = LogEntries{0, 0}
	rf.mu.Lock()
	lastLogIndex := rf.getLastLogIndex()
	for i := 0; i < len(peers); i++ {
		rf.Stat.leaderState.matchIndex[i], rf.Stat.leaderState.nextIndex[i] = 0, lastLogIndex+1
	}
	// start ticker goroutine to start elections
	rf.resetElectionTimer()
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	go rf.ticker()
	go rf.applyCommittedLogs()
	rf.mu.Unlock()
	return rf
}
