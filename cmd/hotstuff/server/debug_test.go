package main

import (
	"context"
	"distributed/consensus"
	"distributed/pb"
	"log"
	"testing"
)

var (
	all = []int32{1, 2, 3, 4}
)

func Debug_Command(command string, targetid ...int32) {
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

func Test_Debug_OutputBlocks(t *testing.T)    { Debug_Command("PrintBlocks", all...) }        // 打印区块链
func Test_Debug_PrintViewNumber(t *testing.T) { Debug_Command("PrintViewNumber") }            // 打印视图号
func Test_Debug_ViewSuccess(t *testing.T)     { Debug_Command("ViewSuccess") }                // 测试视图成功
func Test_Debug_CrossCall(t *testing.T)       { Debug_Command("CrossCall", 2) }               // 测试交叉调用
func Test_Debug_PrintSelfID(t *testing.T)     { Debug_Command("PrintSelfID", all...) }        // 打印自身ID
func Test_Debug_ConnectToOthers(t *testing.T) { Debug_Command("ConnectToOthersandStart", 2) } // 连接其他节点
func Test_Debug_Start(t *testing.T)           { Debug_Command("start") }                      //启动仿真
func Test_Debug_ConnectToSelf(t *testing.T)   { Debug_Command("ConnectToSelf") }              //连接自身
func Test_Debug_pause(t *testing.T)           { Debug_Command("pause") }                      //暂停
func Test_Debug_resume(t *testing.T)          { Debug_Command("resume") }                     //恢复
