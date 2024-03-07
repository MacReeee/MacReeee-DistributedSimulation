package main

import (
	"bufio"
	"context"
	hotstuff "distributed/hotstuff/consensus"
	"distributed/hotstuff/pb"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectToOthers() {
	for i := 1; i <= 4; i++ {
		if i == int(hotstuff.ReplicaID) {
			continue
		}
		hotstuff.NewReplicaClient(int32(i + 1))
	}
}

func main() {
	fmt.Println("在确保所有节点已经启动的情况下，运行控制台程序，按任意键继续...")
	var any string
	fmt.Scanln(&any)

	//创建各个节点的控制台实例
	conns := make([]*grpc.ClientConn, 5)
	conns[1], _ = grpc.Dial(fmt.Sprintf(":%d", 4001), grpc.WithTransportCredentials(insecure.NewCredentials()))
	conns[2], _ = grpc.Dial(fmt.Sprintf(":%d", 4002), grpc.WithTransportCredentials(insecure.NewCredentials()))
	conns[3], _ = grpc.Dial(fmt.Sprintf(":%d", 4003), grpc.WithTransportCredentials(insecure.NewCredentials()))
	conns[4], _ = grpc.Dial(fmt.Sprintf(":%d", 4004), grpc.WithTransportCredentials(insecure.NewCredentials()))
	cons := make([]pb.HotstuffClient, 5)
	cons[1] = pb.NewHotstuffClient(conns[1])
	cons[2] = pb.NewHotstuffClient(conns[2])
	cons[3] = pb.NewHotstuffClient(conns[3])
	cons[4] = pb.NewHotstuffClient(conns[4])

	reader := bufio.NewReader(os.Stdin)
	all := []int32{1, 2, 3, 4}
	for {
		fmt.Println("Command: ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "exit" {
			break
		}

		fmt.Println("请输入作用副本ID: ")
		nodes, _ := reader.ReadString('\n')
		nodes = strings.TrimSpace(nodes)
		var targets []int32
		if nodes == "all" {
			targets = all
		} else {
			nodeIDs := strings.Split(nodes, ",")
			var targets []int
			for _, idStr := range nodeIDs {
				idStr = strings.TrimSpace(idStr) // 去除编号两端的空白字符
				id, err := strconv.Atoi(idStr)
				if err != nil {
					fmt.Printf("无效的节点编号: %s\n", idStr)
					continue
				}
				targets = append(targets, id)
			}
		}
		Command(cmd, targets...)
		fmt.Printf("\n")
	}
}

func Command(command string, targetid ...int32) {
	var targetID []int32
	if len(targetid) == 0 {
		targetID = []int32{1}
	} else {
		targetID = targetid
	}
	for _, id := range targetID {
		client := *hotstuff.NewReplicaClient(id)
		resp, err := client.Debug(context.Background(), &pb.DebugMsg{
			Command: command,
		})
		if resp != nil {
			log.Println("节点", id, "的响应: ", resp.Response)
		} else {
			log.Println("节点", id, "无响应内容")
		}
		log.Println("节点", id, "返回的错误: ", err)
	}
}
