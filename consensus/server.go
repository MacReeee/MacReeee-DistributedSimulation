package hotstuff

import (
	"context"
	d "distributed/dependency"
	"distributed/modules"
	pb2 "distributed/pb"
	"fmt"
	"log"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ReplicaServer) VotePrepare(ctx context.Context, vote *pb2.VoteRequest) (*emptypb.Empty, error) {
	s.WaitForState(Prepare)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)

	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", vote.Voter, " 的Prepare投票")

	//检查类型
	if vote.MsgType != pb2.MsgType_PREPARE_VOTE {
		log.Println("Prepare投票 消息类型不匹配")
		return &emptypb.Empty{}, fmt.Errorf("prepare vote type is not valid")
	}

	//判断视图号
	if vote.ViewNumber < *sync.ViewNumber() {
		log.Println("Prepare投票过旧")
		return &emptypb.Empty{}, fmt.Errorf("prepare vote view number is not valid")
	}

	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		log.Println("prepare投票 签名验证失败")
		return nil, fmt.Errorf("prepare vote signature is not valid")
	}
	count := sync.StoreVote(pb2.MsgType_PREPARE_VOTE, vote)
	once := sync.GetOnce(pb2.MsgType_PREPARE_VOTE)
	if count >= s.threshold { //条件达成，开始执行下一阶段
		log.Println("视图 ", *sync.ViewNumber(), " 的Prepare投票达成阈值")
		sync.TimerReset() //重置计时器
		voters, sigs, _ := sync.GetVoter(pb2.MsgType_PREPARE_VOTE)
		var PreCommitMsg = &pb2.Precommit{}
		once.Do(func() { //调用其他副本的PreCommit
			// 合成QC
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb2.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb2.MsgType_PREPARE_VOTE,
				ViewNumber:   *sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)

			PreCommitMsg = &pb2.Precommit{
				Id:         s.ID,
				MsgType:    pb2.MsgType_PRE_COMMIT,
				ViewNumber: *sync.ViewNumber(),
				Qc:         QC,
				Hash:       vote.Hash,
				Signature:  sig,
				Block:      chain.GetBlock(vote.Hash),
			}

			//模拟投票处理和传输时延
			time.Sleep(d.GetLatency())

			for _, client := range modules.MODULES.ReplicaClient {
				time.Sleep(d.GetProcessTime())
				go (*client).PreCommit(context.Background(), PreCommitMsg)
			}
			s.count = 0 //重置计数器
		})
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) VotePreCommit(ctx context.Context, vote *pb2.VoteRequest) (*emptypb.Empty, error) {
	s.WaitForState(Precommit)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)

	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", vote.Voter, " 的PreCommit投票")
	//检查类型
	if vote.MsgType != pb2.MsgType_PRE_COMMIT_VOTE {
		log.Println("Pre_Commit投票 消息类型不匹配")
		return &emptypb.Empty{}, fmt.Errorf("Pre_Commit vote type is not valid")
	}

	//判断视图号
	if vote.ViewNumber < *sync.ViewNumber() {
		log.Println("Pre_Commit投票过旧")
		return &emptypb.Empty{}, fmt.Errorf("Pre_Commit vote view number is not valid")
	}

	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		log.Println("Pre_Commit Vote 的签名验证失败")
		return nil, fmt.Errorf("Pre_Commit Vote 的签名验证失败")
	}
	count := sync.StoreVote(pb2.MsgType_PRE_COMMIT_VOTE, vote)
	once := sync.GetOnce(pb2.MsgType_PRE_COMMIT_VOTE)
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset() //重置计时器
		voters, sigs, _ := sync.GetVoter(pb2.MsgType_PRE_COMMIT_VOTE)
		var CommitMsg = &pb2.CommitMsg{}
		once.Do(func() { //调用其他副本的Commit
			// 合成QC
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb2.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb2.MsgType_PRE_COMMIT_VOTE,
				ViewNumber:   *sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			CommitMsg = &pb2.CommitMsg{
				Id:         s.ID,
				MsgType:    pb2.MsgType_COMMIT,
				ViewNumber: *sync.ViewNumber(),
				Qc:         QC,
				Hash:       vote.Hash,
				Signature:  sig,
				Block:      chain.GetBlock(vote.Hash),
			}

			//模拟投票处理和传输时延
			time.Sleep(d.GetLatency())

			for _, client := range modules.MODULES.ReplicaClient {
				time.Sleep(d.GetProcessTime())
				go (*client).Commit(context.Background(), CommitMsg)
			}
			s.count = 0 //重置计数器
		})
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) VoteCommit(ctx context.Context, vote *pb2.VoteRequest) (*emptypb.Empty, error) {
	s.WaitForState(Commit)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", vote.Voter, " 的Commit投票")

	//检查类型
	if vote.MsgType != pb2.MsgType_COMMIT_VOTE {
		log.Println("Commit投票 消息类型不匹配")
		return &emptypb.Empty{}, fmt.Errorf("Commit vote type is not valid")
	}

	//判断视图号
	if vote.ViewNumber < *sync.ViewNumber() {
		log.Println("Commit投票过旧")
		return &emptypb.Empty{}, fmt.Errorf("Commit vote view number is not valid")
	}

	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		return nil, fmt.Errorf("commit vote signature is not valid")
	}
	count := sync.StoreVote(pb2.MsgType_COMMIT_VOTE, vote)
	once := sync.GetOnce(pb2.MsgType_COMMIT_VOTE)
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset() //重置计时器
		voters, sigs, _ := sync.GetVoter(pb2.MsgType_COMMIT_VOTE)
		var DecideMsg = &pb2.DecideMsg{}
		once.Do(func() { //调用其他副本的Decide
			// 合成QC
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb2.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb2.MsgType_COMMIT_VOTE,
				ViewNumber:   *sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			DecideMsg = &pb2.DecideMsg{
				Id:         s.ID,
				MsgType:    pb2.MsgType_DECIDE,
				ViewNumber: *sync.ViewNumber(),
				Qc:         QC, //暂未获取
				Hash:       vote.Hash,
				Signature:  sig, //暂未获取
				Block:      chain.GetBlock(vote.Hash),
			}

			//模拟投票处理和传输时延
			time.Sleep(d.GetLatency())

			for _, client := range modules.MODULES.ReplicaClient {
				time.Sleep(d.GetProcessTime())
				go (*client).Decide(context.Background(), DecideMsg)
			}
			s.count = 0 //重置计数器
		})
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) NewView(ctx context.Context, NewViewMsg *pb2.NewViewMsg) (*emptypb.Empty, error) {
	s.WaitForState(Switching)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", NewViewMsg.Voter, " 的NewView消息")

	if NewViewMsg.MsgType != pb2.MsgType_NEW_VIEW {
		log.Println("NewView消息类型不匹配")
		return nil, fmt.Errorf("newview msg type is not valid")
	}

	if sync.GetLeader(NewViewMsg.ViewNumber) != s.ID {
		log.Println("不是当前视图的领导者")
		return nil, fmt.Errorf("not the leader of this view")
	}

	//签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", NewViewMsg.MsgType, NewViewMsg.ViewNumber, NewViewMsg.Hash))

	if NewViewMsg.ViewNumber != 0 && !cryp.Verify(NewViewMsg.Voter, msg, NewViewMsg.Signature) {
		log.Println("NewView消息签名验证失败")
		return nil, fmt.Errorf("newview msg signature is not valid")
	}
	count := sync.StoreVote(pb2.MsgType_NEW_VIEW, nil, NewViewMsg)
	once := sync.GetOnce(pb2.MsgType_NEW_VIEW)
	if count >= s.threshold { //条件达成，开始执行下一阶段
		log.Println("视图 ", *sync.ViewNumber(), " 的NewView消息达成阈值")
		sync.TimerReset() //重置计时器
		var ProposalMsg = &pb2.Proposal{}
		once.Do(func() { //调用其他副本的Propose
			// 获取QC
			HighQC := sync.HighQC()
			// 对QC签名作为自己的签名
			qcjson := QCMarshal(HighQC)
			sig, _ := cryp.NormSign(qcjson)
			// 创建区块
			cmd := []byte("CMD of View: " + strconv.Itoa(int(NewViewMsg.ViewNumber)))
			block := chain.CreateBlock(HighQC.BlockHash, HighQC.ViewNumber+1, HighQC, cmd, s.ID)
			ProposalMsg = &pb2.Proposal{
				Block: block,
				Qc:    HighQC,
				// Aggqc: nil, //hotstuff中用不到
				// ProposalId: 0, //暂未获取，未考虑清楚是否需要该字段
				Proposer:   s.ID,
				ViewNumber: *sync.ViewNumber() + 1,
				Signature:  sig,
				// Timestamp:  0, //暂时用不到
				MsgType: pb2.MsgType_PREPARE,
			}

			//模拟包含区块的传输时延
			time.Sleep(d.GetLatency())

			log.Println("尝试发送视图 ", *sync.ViewNumber()+1, " 的提案 ")

			for _, client := range modules.MODULES.ReplicaClient {
				time.Sleep(d.GetProcessTime())
				go (*client).Prepare(context.Background(), ProposalMsg)
			}

			s.count = 0 //重置计数器
		})
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}
