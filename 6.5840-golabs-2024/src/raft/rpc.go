package raft

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).
	Term         int
	CandidateId  int
	LastLogIndex int // index of candinate's last log entry
	LastLogTerm  int // term of candiante's last log entry
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).
	Term        int  // currentTerm
	VoteGranted bool // true means Candidate received vote
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3A, 3B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// defer DPrintf("{Node %v}'s state is {state %v, term %v} after processing RequestVote,  RequestVoteArgs %v and RequestVoteReply %v ", rf.me, rf.whoami, rf.Stat.currentTerm, args, reply)
	reply.Term, reply.VoteGranted = rf.Stat.currentTerm, false

	if args.Term < rf.Stat.currentTerm { // || (args.Term == rf.Stat.currentTerm && rf.Stat.votedFor != -1 && rf.Stat.votedFor != args.CandidateId) {
		return
	}
	if args.Term > rf.Stat.currentTerm {
		rf.convertToFollower(args.Term)
		// rf.persist()
	}

	lastLogIndex := rf.getLastLogIndex()
	lastLogTerm := rf.getLogTerm(lastLogIndex)
	// if follower's term > candidate's term or
	// follower's last log index/term != candinate's that return false
	// already vote to other candidate
	candidateIsUpToDate := args.LastLogTerm > lastLogTerm ||
		(args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)
	canVote := rf.Stat.votedFor == -1 || rf.Stat.votedFor == args.CandidateId
	if canVote && candidateIsUpToDate {
		rf.Stat.votedFor = args.CandidateId
		rf.persist()
		reply.VoteGranted = true
		rf.resetElectionTimer()
	}
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// 仅重置选举定时器
	reply.Term, reply.Success = rf.Stat.currentTerm, false
	// 拒绝落后任期的节点请求
	if args.Term < rf.Stat.currentTerm {
		return
	}
	rf.resetElectionTimer()
	// 发现更高任期，更改节点为跟随者
	if args.Term > rf.Stat.currentTerm {
		rf.convertToFollower(args.Term)
	}
	// 检查leader之间传递的日志项索引和任期是否满足当前跟随者的要求
	// Follower 日志太短
	if args.PrevLogIndex > rf.getLastLogIndex() {
		reply.ConflictIndex = len(rf.Stat.log)
		reply.ConflictTerm = -1
		return
		// 当跟随者prevLogIndex下的任期和领导者的prevLogTerm都不相等时，需返回冲突的第一个日志索引
	} else if args.PrevLogTerm != rf.getLogTerm(args.PrevLogIndex) {
		reply.ConflictTerm = rf.getLogTerm(args.PrevLogIndex)
		reply.ConflictIndex = args.PrevLogIndex
		// 找到冲突任期的第一个索引，用于加速Leader回退nextIndex
		for i := args.PrevLogIndex - 1; i >= 0; i-- {
			if reply.ConflictTerm != rf.getLogTerm(i) {
				break
			}
			reply.ConflictIndex = i
		}
		return
	}

	// 日志追加与覆盖处理(note: heartbeat still checkout that has conflicted index?)
	conflictIndex := -1
	newEntriesStart := args.PrevLogIndex + 1
	// 检查是否有冲突
	for i, entry := range args.Entries {
		currentIndex := newEntriesStart + i
		if currentIndex >= rf.getLastLogIndex() {
			break // 无冲突，根据prevLogIndex截断进行，并直接追加剩余条目
		}
		if rf.Stat.log[currentIndex].Term != entry.Term {
			// 发现冲突，截断日志到冲突处再追加
			rf.Stat.log = rf.Stat.log[:currentIndex]
			conflictIndex = i
			break
		}

	}
	if conflictIndex != -1 && len(rf.Stat.log) > 0 {
		rf.Stat.log = append(rf.Stat.log, args.Entries...)
		rf.persist()
	} else {
		rf.Stat.log = append(rf.Stat.log[:args.PrevLogIndex+1], args.Entries...)
		rf.persist()
	}
	// 更新提交索引
	if args.LeaderCommit > rf.Stat.committedIndex {
		newCommitIndex := min(args.LeaderCommit, len(rf.Stat.log)-1)
		rf.Stat.committedIndex = newCommitIndex
		rf.applyCond.Broadcast()
		// DPrintf("prevLogIndex:%v,prevLogTerm:%v", args.PrevLogIndex, args.PrevLogTerm)
		// DPrintf("Leader:%v 		sendEntries:%v", args.LeaderId, args.Entries)
		// DPrintf("follower:%v logs:%v ", rf.me, rf.Stat.log)
		// DPrintf("follower Broadcast!")
	}
	reply.Success = true
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntries
	LeaderCommit int
}
type AppendEntriesReply struct {
	Term          int
	Success       bool
	ConflictTerm  int
	ConflictIndex int
}
