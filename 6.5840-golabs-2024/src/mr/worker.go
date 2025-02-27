package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"regexp"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}
type IntermediateFormat struct {
	Key   string
	Value []string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}
func DoMapTask(mapf func(string, string) []KeyValue, reply *TaskAssingnemnt) {
	// 获取文件内容
	content, err := os.ReadFile(reply.Tas.FileName)
	if err != nil {
		log.Fatalf("cannot open %s", reply.Tas.FileName)
	}
	// 调用map函数进行处理
	kva := mapf(reply.Tas.FileName, string(content))

	// 创建保存中间键值对的临时文件
	tempFiles := make([]*os.File, reply.NReduce)
	tempFileNames := make([]string, reply.NReduce)
	//  创建临时文件，并且指定json格式编码器
	encoders := make([]*json.Encoder, reply.NReduce)
	for i := 0; i < reply.NReduce; i++ {
		filename := fmt.Sprintf("mr-%d-%d-tmp-*", reply.Tas.TaskID, i)
		tempFile, err := os.CreateTemp(".", filename)
		if err != nil {
			log.Println("Failed to create temp file:", err)
			// 清理已创建的临时文件
			for j := 0; j < i; j++ {
				os.Remove(tempFileNames[j])
				tempFiles[j].Close()
			}
			return
		}
		// fmt.Printf("%v\n", tempFile.Name())
		tempFileNames[i] = tempFile.Name()
		tempFiles[i] = tempFile
		encoders[i] = json.NewEncoder(tempFile)

	}

	finalStorage := make(map[string][]string)
	for _, kv := range kva {
		finalStorage[kv.Key] = append(finalStorage[kv.Key], kv.Value)

	}
	keys := make([]string, 0, len(finalStorage))
	for k := range finalStorage {
		keys = append(keys, k)
	}
	// 遍历中间键值对，hash(key) % nReduce来确定保存的中间文件索引
	sort.Strings(keys)
	for _, k := range keys {
		index := ihash(k) % reply.NReduce
		encoders[index].Encode(IntermediateFormat{k, finalStorage[k]})
	}
	// 确定最终中间文件名: mr-taskID-PartitionIndex
	for i, file := range tempFiles {
		err := file.Close()
		if err != nil {
			log.Printf("Closing the file fault:%v\n", err)
			// 处理编码错误，关闭所有文件
			for _, file := range tempFiles {
				if file != nil {
					file.Close()
				}
			}
		}
		// 去掉尾部"-tmp"字符串
		newName := fmt.Sprintf("mr-%d-%d", reply.Tas.TaskID, i)
		err = os.Rename(tempFileNames[i], newName)
		if err != nil {
			log.Printf("Rename the file fault:%v\n", err)
		}
	}
}
func DoReduceTask(reducef func(string, []string) string, reply *TaskAssingnemnt) {
	// 设置正则表达式匹配字符
	pattern := fmt.Sprintf(`mr-\d+-%d$`, reply.Tas.TaskID)
	regx, err := regexp.Compile(pattern) //声明正则表达式器
	if err != nil {
		log.Println("regex error!")
		return
	}
	// 获取当前目录的文件
	files, err := os.ReadDir(".")
	if err != nil {
		log.Println("This is not tempfile!")
		return
	}
	// 提前声明统计键值对map[string]string映射表
	finalStorage := make(map[string][]string)
	// 提前创建当前ReduceID输出的最终文件
	tmpFileName := fmt.Sprintf("mr-out-%d-tmp-*", reply.Tas.TaskID)
	ofile, _ := os.CreateTemp(".", tmpFileName)
	defer ofile.Close()
	for _, file := range files {
		// 只有类型为文件，且符合正则表达式匹配的文件才被进入下一步处理
		if !file.IsDir() && regx.MatchString(file.Name()) {
			// fullpath := filepath.Join(".", file.Name())
			fullpath := file.Name()
			f, err := os.Open(fullpath)
			if err != nil {
				log.Println("can't open the file", fullpath)
			}
			defer f.Close()
			// 创建Json解码器
			decoder := json.NewDecoder(f)
			for {
				var kv IntermediateFormat

				if err := decoder.Decode(&kv); err == io.EOF {
					break
				} else if err != nil {
					log.Println("Failed to decode JSON:", err)
					continue
				}
				// 保存读取的中间键值对
				finalStorage[kv.Key] = append(finalStorage[kv.Key], kv.Value...)
			}
		}
	}
	// 准备将中间键有序存储
	keys := make([]string, 0, len(finalStorage))
	for k := range finalStorage {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// 按键有序处理
	for _, key := range keys {
		output := reducef(key, finalStorage[key])
		// 同时有序写入到最终文件中
		fmt.Fprintf(ofile, "%v %v\n", key, output)
	}
	// 为了保持文件的原子性，对临时的最终文件重命名
	newFileName := fmt.Sprintf("mr-out-%d", reply.Tas.TaskID)
	err = os.Rename(ofile.Name(), newFileName)
	if err != nil {
		log.Println("Rename ouput file fault!")
	}
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	// Your worker implementation here.
	for {

		args := RequestTaskTypeArg{}
		reply := TaskAssingnemnt{}
		emptyArgs := EmptyRequestArgs{}
		response := ResponseArgs{}
		ok := call("Coordinator.AssignedTask", &args, &reply)
		// fmt.Println("get task!,the taskID:", reply.Tas.TaskID)
		if !ok {
			fmt.Printf("can't acquire task handle\n")
		}
		switch reply.Tasktype {
		case MapTask:
			fmt.Println("Doing the map task")
			DoMapTask(mapf, &reply)
			response = ResponseArgs{reply.Tas.TaskID}
			call("Coordinator.MarkTaskDone", &response, &emptyArgs)
			// time.Sleep(time.Second)
		case ReduceTask:
			fmt.Println("Doing the reduce task")
			DoReduceTask(reducef, &reply)
			response = ResponseArgs{reply.Tas.TaskID}
			call("Coordinator.MarkTaskDone", &response, &emptyArgs)
			// time.Sleep(time.Second)
		case WaitTask:
			// fmt.Println("the task is waitting now")
			time.Sleep(time.Second)
		case Finished:
			// fmt.Println("the task is exitting now")
			return
		default:
			fmt.Printf("the task type: %v not found!\n", reply)
			continue
		}
	}

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
// func CallExample() {

// 	// declare an argument structure.
// 	args := RequestArgs{}

// 	// declare a reply structure.
// 	reply := TaskAssingnemnt{}

// 	// send the RPC request, wait for the reply.
// 	// the "Coordinator.Example" tells the
// 	// receiving server that we'd like to call
// 	// the Example() method of struct Coordinator.
// 	ok := call("Coordinator.AssignedTask", &args, &reply)
// 	if ok {
// 		fmt.Printf("reply.Y %v\n", reply.Y)
// 	} else {
// 		fmt.Printf("call failed!\n")
// 	}
// }

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
		// log.Println("Coordinator is down!, the worker will be exit!")
		// os.Exit(0)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
