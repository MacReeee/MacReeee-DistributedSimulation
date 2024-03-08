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

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("输入要调试节点, 默认1, 全部输入 'all'")
	nodes, _ := reader.ReadString('\n')
	nodes = strings.TrimSpace(nodes)
	var tar []int32
	if nodes == "all" {
		tar = []int32{1, 2, 3, 4}
	} else {
		tar = []int32{1}
	}

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

	all := []int32{1, 2, 3, 4}
	for {
		fmt.Println("Command: ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "exit" {
			break
		}

		if cmd == "restart" {
			conns[1], _ = grpc.Dial(fmt.Sprintf(":%d", 4001), grpc.WithTransportCredentials(insecure.NewCredentials()))
			conns[2], _ = grpc.Dial(fmt.Sprintf(":%d", 4002), grpc.WithTransportCredentials(insecure.NewCredentials()))
			conns[3], _ = grpc.Dial(fmt.Sprintf(":%d", 4003), grpc.WithTransportCredentials(insecure.NewCredentials()))
			conns[4], _ = grpc.Dial(fmt.Sprintf(":%d", 4004), grpc.WithTransportCredentials(insecure.NewCredentials()))
			cons[1] = pb.NewHotstuffClient(conns[1])
			cons[2] = pb.NewHotstuffClient(conns[2])
			cons[3] = pb.NewHotstuffClient(conns[3])
			cons[4] = pb.NewHotstuffClient(conns[4])
			fmt.Println("重连成功")
			continue
		}

		if len(tar) == 1 {
			Command(cmd, 1)
		} else {
			fmt.Println("请输入作用副本ID: ")
			nodes, _ := reader.ReadString('\n')
			nodes = strings.TrimSpace(nodes)
			var targets []int32
			if (nodes == "all") || (nodes == "") {
				targets = all
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
			Command(cmd, targets...)
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
