package hotstuff

import (
	"context"
	d "distributed/dependency"
	"distributed/modules"
	pb2 "distributed/pb"
	"fmt"
	"log"
	stsync "sync"
	"time"
)

var (
	wg       stsync.WaitGroup //中断控制
	StopFlag = false          //中断标志
)

func (s *ReplicaServer) Debug(ctx context.Context, debug *pb2.DebugMsg) (*pb2.DebugMsg, error) {
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	fmt.Printf("\n接收到调试指令: %v\n", debug.Command)
	switch debug.Command {
	////打印区块链
	//case "PrintBlocks":
	//	PrintChain()
	//	return &pb2.DebugMsg{}, nil
	// 打印视图号
	case "PrintViewNumber":
		log.Println("当前视图号: ", *sync.ViewNumber())
		return &pb2.DebugMsg{}, nil
	//启动仿真程序
	case "StartAll", "sa", "启动":
		sync.Start()
		if s.ID == 1 {
			s.SetState(Switching)
			highQC := s.PrepareQC
			qcjson := QCMarshal(highQC)
			sig, _ := cryp.NormSign(qcjson)
			cmd := []byte("CMD of View: 1")
			block := chain.CreateBlock([]byte("FFFFFFFFFFFF"), 1, highQC, cmd, 1)
			var ProposalMsg = &pb2.Proposal{
				Block:      block,
				Qc:         highQC,
				Proposer:   1,
				ViewNumber: *sync.ViewNumber() + 1,
				Signature:  sig,
				MsgType:    pb2.MsgType_PREPARE,
			}
			clients := modules.MODULES.ReplicaClient
			fmt.Println("等待", d.Configs.BuildInfo.NumReplicas/2, "秒后开始发送提案...")
			time.Sleep(500 * time.Duration(d.Configs.BuildInfo.NumReplicas) / 2 * time.Millisecond) //防止节点过多时还有节点未接收到启动命令
			for _, client := range clients {
				go (*client).Prepare(context.Background(), ProposalMsg)
			}
			return &pb2.DebugMsg{Response: "已执行启动程序"}, nil
		}
		s.SetState(Switching)
		return &pb2.DebugMsg{}, nil
	//控制节点连接其他节点
	case "ConnectToOthers", "cto", "原神":
		var nums, i int32
		nums = d.Configs.BuildInfo.NumReplicas
		for i = 1; i <= nums; i++ {
			NewReplicaClient(i)
		}
		return &pb2.DebugMsg{}, nil
	case "TestTimeoutStart", "tts":
		log.Printf("启动视图计时器，等待超时\n")
		s.SetState(Switching)
		sync.Start()
		return &pb2.DebugMsg{}, nil
	case "ViewNumpp", "vpp":
		sync.ViewNumberPP()
		log.Println("视图号+1")
		return &pb2.DebugMsg{}, nil
	case "PrintSelfID":
		log.Println("当前节点ID: ", s.ID)
		return &pb2.DebugMsg{}, nil
	case "ConnectToSelf":
		NewReplicaClient(s.ID)
		return &pb2.DebugMsg{Response: "执行完毕"}, nil
	case "pause":
		StopFlag = true
		wg.Add(1)
		return &pb2.DebugMsg{Response: "已暂停仿真"}, nil
	case "testLatency", "tl":
		log.Println("当前传输时延抽样: ", d.GetLatency())
		log.Println("当前处理时延抽样: ", d.GetProcessTime())
		return &pb2.DebugMsg{}, nil
	case "resume":
		StopFlag = false
		wg.Done()
		return &pb2.DebugMsg{Response: "已恢复仿真"}, nil
	case "clc":
		fmt.Printf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
		return &pb2.DebugMsg{}, nil
	case "debug":
		BuildInfo := d.Configs.BuildInfo
		BuildInfo.DebugMode = !BuildInfo.DebugMode
		log.Println("当前Debug状态: ", BuildInfo.DebugMode)
		return &pb2.DebugMsg{}, nil
	case "syncinfo":
		s := modules.MODULES.Synchronizer
		log.Println("当前同步信息: ", s)
		return &pb2.DebugMsg{}, nil
	case "load":
		d.LoadFromFile()
		return &pb2.DebugMsg{}, nil
	case "PrintConfig":
		d.ReadConfig()
		return &pb2.DebugMsg{}, nil
	case "reset", "r":
		modules.MODULES.Reset <- true
		return &pb2.DebugMsg{}, nil
	default:
		log.Println("未知的调试命令...")
		return &pb2.DebugMsg{Response: "未知的调试命令: " + debug.Command}, nil
	}
}

func (s *ReplicaServer) SelfID() int32 {
	return s.ID
}
