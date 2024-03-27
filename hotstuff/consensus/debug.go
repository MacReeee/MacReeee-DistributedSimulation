package hotstuff

import (
	"context"
	d "distributed/hotstuff/dependency"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"fmt"
	"log"
	stsync "sync"
)

var (
	wg       stsync.WaitGroup //中断控制
	StopFlag = false          //中断标志
)

func (s *ReplicaServer) Debug(ctx context.Context, debug *pb.DebugMsg) (*pb.DebugMsg, error) {
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	fmt.Printf("\n接收到调试指令: %v\n", debug.Command)
	switch debug.Command {
	//打印区块链
	case "PrintBlocks":
		PrintChain()
		return &pb.DebugMsg{}, nil
	// 打印视图号
	case "PrintViewNumber":
		log.Println("当前视图号: ", *sync.ViewNumber())
		return &pb.DebugMsg{}, nil
	//启动仿真程序
	case "StartAll", "sa":
		if s.ID == 1 {
			highQC := s.PrepareQC
			qcjson := QCMarshal(highQC)
			sig, _ := cryp.NormSign(qcjson)
			cmd := []byte("CMD of View: 1")
			block := chain.CreateBlock([]byte("FFFFFFFFFFFF"), 1, highQC, cmd, 1)
			var ProposalMsg = &pb.Proposal{
				Block:      block,
				Qc:         highQC,
				Proposer:   1,
				ViewNumber: *sync.ViewNumber() + 1,
				Signature:  sig,
				MsgType:    pb.MsgType_PREPARE,
			}
			clients := modules.MODULES.ReplicaClient
			for _, client := range clients {
				go (*client).Prepare(context.Background(), ProposalMsg)
			}
			return &pb.DebugMsg{Response: "已执行启动程序"}, nil
		}
		return &pb.DebugMsg{}, nil
	//控制节点连接其他节点
	case "ConnectToOthers", "cto":
		var nums int
		if d.DebugMode {
			nums = 4
		} else {
			nums = 4
		}
		for i := 1; i <= nums; i++ {
			NewReplicaClient(int32(i))
		}
		sync.Start()
		return &pb.DebugMsg{}, nil
	case "PrintSelfID":
		log.Println("当前节点ID: ", s.ID)
		return &pb.DebugMsg{}, nil
	case "ConnectToSelf":
		NewReplicaClient(s.ID)
		return &pb.DebugMsg{Response: "执行完毕"}, nil
	case "pause":
		StopFlag = true
		wg.Add(1)
		return &pb.DebugMsg{Response: "已暂停仿真"}, nil
	case "resume":
		StopFlag = false
		wg.Done()
		return &pb.DebugMsg{Response: "已恢复仿真"}, nil
	case "clc":
		fmt.Printf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
		return &pb.DebugMsg{}, nil
	case "debug":
		d.DebugMode = !d.DebugMode
		log.Println("当前Debug状态: ", d.DebugMode)
		return &pb.DebugMsg{}, nil
	case "syncinfo":
		s := modules.MODULES.Synchronizer
		log.Println("当前同步信息: ", s)
		return &pb.DebugMsg{}, nil
	case "load":
		d.LoadFromFile()
		return &pb.DebugMsg{}, nil
	case "PrintConfig":
		d.ReadConfig()
		return &pb.DebugMsg{}, nil
	default:
		log.Println("未知的调试命令...")
		return &pb.DebugMsg{Response: "未知的调试命令: " + debug.Command}, nil
	}
}

func PrintChain() {
	var (
		// sync  = modules.MODULES.Synchronizer
		// cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	_, _, keys := chain.GetBlockChain()

	var i = 0
	for _, block := range keys {
		fmt.Printf("\n区块 %d 的信息:\n", i)

		fmt.Println("区块Hash:\t", string(block.Hash))
		fmt.Println("父Hash:\t\t", string(block.ParentHash))
		fmt.Println("区块高度:\t", block.Height)
		fmt.Println("子区块的Hash:\t", block.Children)
		fmt.Println("区块的内容:\t", string(block.Cmd), "\n")
		i++
	}
}

func (s *ReplicaServer) SelfID() int32 {
	return s.ID
}
