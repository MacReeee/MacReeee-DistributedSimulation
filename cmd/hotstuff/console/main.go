package main

import (
	"bufio"
	"context"
	"distributed/consensus"
	pb2 "distributed/pb"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("输入节点数量: ")
	numStr, _ := reader.ReadString('\n')
	numStr = strings.TrimSpace(numStr)
	num, _ := strconv.Atoi(numStr)

	fmt.Println("输入要调试节点, 默认1, 全部输入 'all'")
	nodes, _ := reader.ReadString('\n')
	nodes = strings.TrimSpace(nodes)
	var tar []int32
	if nodes == "all" {
		tar = make([]int32, num)
		for i := 1; i <= num; i++ {
			tar[i-1] = int32(i)
		}
	} else if nodes == "" {
		tar = []int32{1}
	} else {
		nodeIDs := strings.Split(nodes, " ")
		for _, idStr := range nodeIDs {
			idStr = strings.TrimSpace(idStr)
			id, err := strconv.Atoi(idStr)
			if err != nil {
				fmt.Printf("无效的节点编号: %s\n", idStr)
				continue
			}
			tar = append(tar, int32(id))
		}
	}

	all := make([]int32, num)
	for i := 1; i <= num; i++ {
		all[i-1] = int32(i)
	}
	for {
		fmt.Println("Command: ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "exit" {
			break
		}

		if len(tar) == 1 {
			Command(cmd, 1)
		} else if len(tar) == num {
			Command(cmd, all...)
		} else {
			fmt.Println("请输入作用副本ID: ")
			nodes, _ := reader.ReadString('\n')
			nodes = strings.TrimSpace(nodes)
			var targets []int32
			if (nodes == "all") || (nodes == "") {
				targets = tar
			} else {
				nodeIDs := strings.Split(nodes, " ")
				for _, idStr := range nodeIDs {
					idStr = strings.TrimSpace(idStr)
					id, err := strconv.Atoi(idStr)
					if err != nil {
						fmt.Printf("无效的节点编号: %s\n", idStr)
						continue
					}
					targets = append(targets, int32(id))
				}
			}
			if cmd == "sa30" {
				Command("sa", targets...)
				fmt.Println("等待30s...")
				<-time.After(30 * time.Second)
				Command("pause", targets...)
			} else {
				Command(cmd, targets...)
			}
		}
		fmt.Printf("\n")
	}
}

var commandList = []string{
	"OutputBlocks", "PrintViewNumber", "ViewSuccess", "CrossCall",
	"PrintSelfID", "ConnectToOthers", "Start",
}

func Command(command string, targetid ...int32) {
	var targetID []int32
	if len(targetid) == 0 {
		targetID = []int32{1}
	} else {
		targetID = targetid
	}
	wg := sync.WaitGroup{}
	wg.Add(len(targetID))
	for _, id := range targetID {
		go func(id int32) {
			client := *hotstuff.NewReplicaClient(id)
			resp, err := client.Debug(context.Background(), &pb2.DebugMsg{
				Command: command,
			})
			if resp != nil {
				log.Println("节点", id, "的响应: ", resp.Response)
			}
			if err != nil {
				log.Println("节点", id, "返回的错误: ", err)
			}
			wg.Done()
		}(id)
	}
	wg.Wait()
}
