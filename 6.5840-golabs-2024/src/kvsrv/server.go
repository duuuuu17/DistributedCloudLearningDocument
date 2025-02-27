package kvsrv

import (
	"log"
	"sync"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type KVServer struct {
	mu sync.Mutex
	// Your definitions here.
	value   map[string]string
	lastSeq map[int64]int64 // ClientID -> OpSeq
	// record  map[int64]map[int64]string // ClientID -> (OpSeq -> value)
	record map[int64]string
}

// 在线性一致性下，get的操作永远是最新的值
func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if exsits, ok := kv.value[args.Key]; ok {
		reply.Value = exsits
	} else {
		reply.Value = ""
	}

}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	// 当请求序列号小于等于服务器记录的ClientID的已处理的序列号时，返回记录的value或空字符串(表示不执行)
	if kv.lastSeq[args.ClientID] >= args.Seq {
		reply.Value = kv.record[args.ClientID]
		return
	}
	// 当请求的序列号远大于服务器记录的ClientID的已处理的序列号时，表示该操作不执行
	if kv.lastSeq[args.ClientID]+1 != args.Seq {
		reply.Value = ""
		return
	}
	// 当args.Seq == kv.lastSeq[args.ClientID]+1 时，表明该操作是下一个要执行的操作
	reply.Value = args.Value
	kv.lastSeq[args.ClientID] = args.Seq
	kv.record[args.ClientID] = args.Value
	kv.value[args.Key] = args.Value

}

// 除了key的赋值value时的逻辑不同，其余判断序列号的逻辑相同
func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.lastSeq[args.ClientID] >= args.Seq {
		reply.Value = kv.record[args.ClientID]
		return
	}
	if kv.lastSeq[args.ClientID]+1 != args.Seq {
		reply.Value = ""
		return
	}

	kv.lastSeq[args.ClientID] = args.Seq
	if exsits, ok := kv.value[args.Key]; ok { //如果存在
		kv.record[args.ClientID] = exsits
		reply.Value = exsits
		kv.value[args.Key] = exsits + args.Value
	} else {
		// 如果不存在
		kv.record[args.ClientID] = ""
		reply.Value = ""
		kv.value[args.Key] = args.Value
	}
}

// 用于快速清理已处理的序列号暂存结果
func (kv *KVServer) Delete(args *DeleteArgs, reply *DeleteReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.record, args.ClientID)

}

func StartKVServer() *KVServer {
	kv := new(KVServer)
	// You may need initialization code here.
	kv.mu = sync.Mutex{}

	kv.value = make(map[string]string)
	kv.record = make(map[int64]string)
	kv.lastSeq = make(map[int64]int64)

	return kv
}
