package hotstuff

import (
	"context"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"fmt"
	"log"
	"time"
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
	case "OutputBlocks":
		chainByHash, chainByHeight := chain.GetBlockChain()
		log.Println("哈希: ", chainByHash)
		log.Println("高度: ", chainByHeight)
		return &pb.DebugMsg{}, nil
	// 打印视图号
	case "PrintViewNumber":
		log.Println("当前视图号: ", *sync.ViewNumber())
		return &pb.DebugMsg{}, nil
	//测试视图成功功能
	case "ViewSuccess":
		_, success := sync.GetContext()
		log.Println("当前视图号: ", sync.ViewNumber())
		sync.Start()
		success()
		log.Println("调用ViewSuccess后的视图号: ", sync.ViewNumber())
		go func() {
			time.Sleep(15 * time.Second)
			log.Println("15s后的视图号: ", sync.ViewNumber())
		}()
		return &pb.DebugMsg{}, nil
	//启动仿真程序
	case "start":
		sync.Start()
		if s.ID == 1 {
			highQC := s.PrepareQC
			qcjson := QCMarshal(highQC)
			sig, _ := cryp.NormSign(qcjson)
			cmd := []byte("CMD of View: 1")
			block := chain.CreateBlock([]byte("FFFFFFFFFFFF"), 0, highQC, cmd, 1)
			var ProposalMsg = &pb.Proposal{
				Block:      block,
				Qc:         highQC,
				Proposer:   1,
				ViewNumber: *sync.ViewNumber(),
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
	case "ConnectToOthers":
		for i := 1; i <= 4; i++ {
			NewReplicaClient(int32(i))
		}
		return &pb.DebugMsg{}, nil
	case "原神，启动！", "ConnectToSelfandStart":
		NewReplicaClient(s.ID)
		sync.Start()
		if s.ID == 1 {
			highQC := s.PrepareQC
			qcjson := QCMarshal(highQC)
			sig, _ := cryp.NormSign(qcjson)
			cmd := []byte("CMD of View: 1")
			block := chain.CreateBlock([]byte("FFFFFFFFFFFF"), 0, highQC, cmd, 1)
			var ProposalMsg = &pb.Proposal{
				Block:      block,
				Qc:         highQC,
				Proposer:   1,
				ViewNumber: *sync.ViewNumber(),
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
	//测试节点之间的相互调用
	case "CrossCall":
		clients := modules.MODULES.ReplicaClient
		// 初始化一个默认的成功响应
		response := &pb.DebugMsg{Response: "所有副本调用成功"}

		for i := 1; i <= 4; i++ {
			client := clients[int32(i)]
			if client != nil {
				_, err := (*client).Debug(context.Background(), &pb.DebugMsg{Command: "PrintSelfID"})
				if err != nil {
					// 处理每个client调用的错误，这里选择了记录错误并修改响应消息
					fmt.Printf("调用副本 %d 失败: %v\n", i, err)
					response.Response = "至少一个副本调用失败"
					return response, err
				}
			} else {
				fmt.Printf("副本 %d 未连接\n", i)
				return &pb.DebugMsg{Response: "侦测到网络未初始化"}, fmt.Errorf("副本 %d 未连接", i)
			}
		}
		return response, nil
	case "PrintSelfID":
		log.Println("当前节点ID: ", s.ID)
		return &pb.DebugMsg{}, nil
	case "ConnectToSelf":
		NewReplicaClient(s.ID)
		return &pb.DebugMsg{Response: "执行完毕"}, nil
	case "pause":
		StopFlag = true
		return &pb.DebugMsg{Response: "已暂停仿真"}, nil
	case "resume":
		StopFlag = false
		wg.Done()
		return &pb.DebugMsg{Response: "已恢复仿真"}, nil
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
	chainByHash, _ := chain.GetBlockChain()
	for hash, block := range chainByHash {
		fmt.Println("区块Hash: ", hash, "\n", "子区块的Hash: ", block.Children, "\n", "父Hash: ", string(block.ParentHash))
	}
}
