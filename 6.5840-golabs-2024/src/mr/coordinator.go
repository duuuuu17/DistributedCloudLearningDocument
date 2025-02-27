package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	// Your definitions here.
	File                       []string
	nReduce                    int
	nCurrentAssignedReduceTask int
	nCurrentAssignedMapTask    int
	phase                      SchedulePahse
	tasks                      []Task
	rePendingTask              []Task
	mtx                        sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.

// 分配任务
func (c *Coordinator) AssignedTask(arg *RequestTaskTypeArg, replyAssign *TaskAssingnemnt) error {

	// fmt.Printf("the currentPhase:%v\n", c.phase.currentPhase)

	c.Check()

	c.mtx.Lock()
	defer c.mtx.Unlock()
	switch c.phase {
	// 执行map阶段任务处理
	case MapPhase:
		if len(c.File) != 0 {
			replyAssign.Tas = Task{c.File[0],
				c.nCurrentAssignedMapTask,
				time.Now(),
				Doing}
			replyAssign.NReduce = c.nReduce
			replyAssign.Tasktype = MapTask
			c.tasks = append(c.tasks, replyAssign.Tas)
			c.File = c.File[1:]
			c.nCurrentAssignedMapTask++
		} else if len(c.rePendingTask) > 0 {
			fmt.Println("assigned repending map task!")
			replyAssign.Tas = c.rePendingTask[0]
			c.tasks[replyAssign.Tas.TaskID].StartTime = time.Now()
			replyAssign.Tas.StartTime = time.Now()
			replyAssign.Tas.Status = Doing
			replyAssign.NReduce = c.nReduce
			replyAssign.Tasktype = MapTask
			c.rePendingTask = c.rePendingTask[1:]
		} else {
			c.HandleTimeOutTask()
			fmt.Println("wait map task!")
			replyAssign.Tasktype = WaitTask
		}
	// 执行Reduce阶段任务返回
	case ReducePhase:
		if c.nCurrentAssignedReduceTask != c.nReduce {
			replyAssign.Tas = Task{
				FileName:  "", // reduce处理的中间文件只需要通过os.ReadDir()获取+正则表达式器筛选
				TaskID:    c.nCurrentAssignedReduceTask,
				StartTime: time.Now(),
				Status:    Doing,
			}
			replyAssign.Tasktype = ReduceTask
			c.nCurrentAssignedReduceTask++
			c.tasks = append(c.tasks, replyAssign.Tas)
		} else if len(c.rePendingTask) > 0 {
			fmt.Println("assigned repending map task!")
			replyAssign.Tas = c.rePendingTask[0]
			c.tasks[replyAssign.Tas.TaskID].StartTime = time.Now()
			replyAssign.Tas.StartTime = time.Now()
			replyAssign.Tas.Status = Doing
			replyAssign.Tasktype = ReduceTask
			c.rePendingTask = c.rePendingTask[1:]
		} else {
			c.HandleTimeOutTask()
			fmt.Println("wait reduce task!")
			replyAssign.Tasktype = WaitTask
		}
	case FinishPhase:
		replyAssign.Tasktype = Finished
	}
	return nil
}

// 当有任务完成时，将task标记为完成状态
func (c *Coordinator) MarkTaskDone(arg *ResponseArgs, emarg *EmptyRequestArgs) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	// fmt.Println("had task done!,the taskID:", arg.TaskID)
	c.tasks[arg.TaskID].Status = Completed
	return nil
}
func (c *Coordinator) HandleTimeOutTask() {
	const outTime = 10 * time.Second
	if c.phase == FinishPhase {
		return
	}
	for _, task := range c.tasks {
		if task.Status == Doing && time.Since(task.StartTime) > outTime {
			fmt.Println("some task outTime!")
			task.StartTime = time.Now()
			task.Status = Pending
			c.rePendingTask = append(c.rePendingTask, task)
			return
		}
	}

}
func (c *Coordinator) AllTaskDone() bool {

	if len(c.tasks) == 0 {
		return false
	}
	for _, task := range c.tasks {
		if task.Status != Completed {
			return false
		}
	}
	return true
}
func (c *Coordinator) Check() bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	switch c.phase {
	case MapPhase:
		if c.AllTaskDone() && len(c.File) == 0 {
			if len(c.rePendingTask) == 0 {
				fmt.Println("All maptask are done!")
				c.phase = ReducePhase
				c.tasks = []Task{}
			}
		}
	case ReducePhase:
		if c.AllTaskDone() && c.nCurrentAssignedReduceTask == c.nReduce {
			if len(c.rePendingTask) == 0 {
				fmt.Println("All reducetask are done!")
				c.phase = FinishPhase
			}
		}
	}
	return true
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() { // 将coordinator结构注册到rpc中
	rpc.Register(c)
	rpc.HandleHTTP()
	// l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	// Your code here.
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.phase == FinishPhase {
		time.Sleep(3 * time.Second)
		return true
	}
	return false
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		File:                       files,
		nReduce:                    nReduce,
		nCurrentAssignedMapTask:    0,
		nCurrentAssignedReduceTask: 0,
		phase:                      MapPhase,
		tasks:                      []Task{},
		rePendingTask:              []Task{},
		mtx:                        sync.Mutex{},
	}
	// Your code here.
	c.server()
	return &c
}
