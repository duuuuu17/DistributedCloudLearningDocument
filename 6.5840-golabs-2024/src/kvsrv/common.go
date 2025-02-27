package kvsrv

// Put or Append
type PutAppendArgs struct {
	Key   string
	Value string
	// You'll have to add definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Seq      int64
	ClientID int64
}

type PutAppendReply struct {
	Value string
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
}
type DeleteArgs struct {
	ClientID int64
}
type DeleteReply struct {
}
type GetReply struct {
	Value string
}

const (
	ErrVersion string = "Error Version"
	ErrNoKey   string = "Error NoKey"
)
