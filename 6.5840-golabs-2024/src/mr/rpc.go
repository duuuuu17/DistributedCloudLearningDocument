package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//
import (
	"os"
	"strconv"
	"time"
)

type Task struct {
	FileName  string
	TaskID    int
	StartTime time.Time
	Status    TaskStatus
}
type EmptyRequestArgs struct{}
type RequestArgs struct {
	RequestTaskType TaskType
	Task_           Task
	// workerID        int
}
type PhaseValue struct {
	Value int8
}
type RequestTaskTypeArg struct {
	RequestTaskType TaskType
	// workerID        int
}

// 任务分配的task
type TaskAssingnemnt struct {
	Tas      Task
	NReduce  int
	Tasktype TaskType // 第二次修改设计时使用
}

type Reply struct {
	// Tasktype TaskType // 第二次修改设计时使用
	TaskID int // 完成的任务ID
	Status TaskStatus
}

// heart check
type ResponseArgs struct {
	// TaskType TaskType
	TaskID int
}
type SchedulePahse int8

const (
	MapPhase SchedulePahse = iota
	ReducePhase
	FinishPhase
)

type TaskType string

const (
	MapTask    TaskType = "map"
	ReduceTask TaskType = "reduce"
	WaitTask   TaskType = "wait"
	Finished   TaskType = "finished"
)

type TaskStatus string

const (
	Completed TaskStatus = "completed"
	Doing     TaskStatus = "Doing"
	Pending   TaskStatus = "Pending"
	Error     TaskStatus = "Error"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

// type ResultCode int

// const (
// 	ResultError       ResultCode = -1
// 	ResultSuccess     ResultCode = 1
// 	ResultConnFailure ResultCode = 0
// )

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
