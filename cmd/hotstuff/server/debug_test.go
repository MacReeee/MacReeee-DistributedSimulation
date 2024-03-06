package main

import (
	"context"
	hotstuff "distributed/hotstuff/consensus"
	"distributed/hotstuff/pb"
	"log"
	"testing"
)

// 打印区块链
func Test_Debug_OutputBlocks(t *testing.T) {
	client := *hotstuff.NewReplicaClient(1)
	_, err := client.Debug(context.Background(), &pb.DebugMsg{
		Command: "OutputBlocks",
	})
	log.Println(err)
}

// 打印视图号
func Test_Debug_PrintViewNumber(t *testing.T) {
	client := *hotstuff.NewReplicaClient(1)
	_, err := client.Debug(context.Background(), &pb.DebugMsg{
		Command: "PrintViewNumber",
	})
	log.Println(err)
}

// 测试视图成功
func Test_Debug_ViewSuccess(t *testing.T) {
	client := *hotstuff.NewReplicaClient(1)
	_, err := client.Debug(context.Background(), &pb.DebugMsg{
		Command: "ViewSuccess",
	})
	log.Println(err)
}
