package hotstuff

import (
	"context"
	d "distributed/dependency"
	"distributed/modules"
	pb2 "distributed/pb"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"time"
)

func (s *ReplicaServer) Prepare(ctx context.Context, Proposal *pb2.Proposal) (*emptypb.Empty, error) {
	//用以控制台控制中断
	if StopFlag {
		wg.Wait()
	}
	s.WaitForState(Switching)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)

	fmt.Printf("\n")
	log.Println("接收到来自视图 ", Proposal.ViewNumber, "中节点 ", Proposal.Proposer, " 的提案，等待验证后进入新视图，当前视图: ", *sync.ViewNumber())

	// 3. 确保区块来自主节点
	if Proposal.Proposer != sync.GetLeader(Proposal.ViewNumber) {
		log.Println("提案的提议者不是当前视图的领导者")
		s.SetState(Switching)
		return nil, fmt.Errorf("proposer is not the leader of current view")
	}

	// 4. 检查提案是否满足接收条件
	if !s.VoteRule(Proposal) {
		log.Println("提案不安全")
		return nil, fmt.Errorf("proposal is not valid")
	}

	// 5. 检查区块的视图号/高度是否满足同一高度仅能投票一次
	if Proposal.ViewNumber <= s.lastVote {
		log.Println("接收到的区块视图号小于上一次投票的视图号，拒绝投票")
		return nil, fmt.Errorf("proposal is not valid")
	}

	// 切换到合适视图
	if *sync.ViewNumber() < Proposal.ViewNumber {
		sync.ViewNumberSet(Proposal.ViewNumber)
		s.TempViewNumber = Proposal.ViewNumber
		once := sync.GetOnly()
		once.Do(sync.Success)
	}

	s.lastVote = Proposal.ViewNumber

	s.SetState(Prepare)

	// 临时存储区块
	chain.Store(Proposal.Block)

	//对提案进行签名
	sig, err := cryp.Sign(pb2.MsgType_PREPARE_VOTE, *sync.ViewNumber(), Proposal.Block.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PrepareVoteMsg := &pb2.VoteRequest{
		// ProposalId: Proposal.ProposalId,
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       Proposal.Block.Hash,
		MsgType:    pb2.MsgType_PREPARE_VOTE,
	}

	var leader pb2.HotstuffClient
	leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]

	//log.Println("视图 ", *sync.ViewNumber(), " 的领导者是: ", sync.GetLeader())

	//模拟投票处理和传输时延
	time.Sleep(d.GetLatency())
	if !d.DenyVote {
		go leader.VotePrepare(context.Background(), PrepareVoteMsg)
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) PreCommit(ctx context.Context, PrecommitMsg *pb2.Precommit) (*emptypb.Empty, error) {
	s.WaitForState(Prepare)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)

	//确保消息来自主节点
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", PrecommitMsg.Id, " 的PreCommit消息")
	if PrecommitMsg.Id != sync.GetLeader(PrecommitMsg.ViewNumber) {
		log.Println("Precommit消息的提议者不是当前视图的领导者")
		return nil, fmt.Errorf("PrecommitMsg is not valid")
	}

	//检查提案是否满足接收条件
	if !s.VoteRulePreCommit(PrecommitMsg) {
		log.Println("Precommit消息不安全")
		return nil, fmt.Errorf("PrecommitMsg is not valid")
	}

	//检查区块的视图号/高度是否满足同一高度仅能投票一次
	if PrecommitMsg.ViewNumber < s.lastVote {
		log.Println("PreCommit 消息的视图号小于上一次投票的视图号，拒绝投票")
		return nil, fmt.Errorf("PreCommit Msg is too old")
	}

	// 切换到合适视图
	if *sync.ViewNumber() < PrecommitMsg.ViewNumber {
		sync.ViewNumberSet(PrecommitMsg.ViewNumber)
		s.TempViewNumber = PrecommitMsg.ViewNumber
		once := sync.GetOnly()
		once.Do(sync.Success)
	}

	s.lastVote = PrecommitMsg.ViewNumber

	//if ok, err := MatchingMsg(PrecommitMsg.MsgType, PrecommitMsg.ViewNumber, pb2.MsgType_PRE_COMMIT, *sync.ViewNumber()); !ok {
	//	return nil, err
	//}

	s.PrepareQC = PrecommitMsg.Qc //更新PrepareQC

	sync.TimerReset()

	s.SetState(Precommit)

	block := PrecommitMsg.Block
	if chain.GetBlock(block.Hash) == nil {
		chain.Store(block) // 存储这一轮的区块
	}

	sig, err := cryp.Sign(pb2.MsgType_PRE_COMMIT_VOTE, *sync.ViewNumber(), PrecommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PreCommitVoteMsg := &pb2.VoteRequest{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       PrecommitMsg.Hash,
		MsgType:    pb2.MsgType_PRE_COMMIT_VOTE,
	}
	var leader pb2.HotstuffClient
	leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]

	//模拟投票处理和传输时延
	time.Sleep(d.GetLatency())

	if !d.DenyVote {
		go leader.VotePreCommit(context.Background(), PreCommitVoteMsg)
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Commit(ctx context.Context, CommitMsg *pb2.CommitMsg) (*emptypb.Empty, error) {
	s.WaitForState(Precommit)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", CommitMsg.Id, " 的Commit消息")

	//确保消息来自主节点
	if CommitMsg.Id != sync.GetLeader(CommitMsg.ViewNumber) {
		log.Println("Commit消息的提议者不是当前视图的领导者")
		return nil, fmt.Errorf("CommitMsg is not valid")
	}

	//检查提案是否满足接收条件
	if !s.VoteRuleCommit(CommitMsg) {
		log.Println("Commit消息不安全")
		return nil, fmt.Errorf("CommitMsg is not valid")
	}

	//检查区块的视图号/高度是否满足同一高度仅能投票一次
	if CommitMsg.ViewNumber < s.lastVote {
		log.Println("CommitMsg 消息的视图号小于上一次投票的视图号，拒绝投票")
		return nil, fmt.Errorf("CommitMsg Msg is too old")
	}

	// 切换到合适视图
	if *sync.ViewNumber() < CommitMsg.ViewNumber {
		sync.ViewNumberSet(CommitMsg.ViewNumber)
		s.TempViewNumber = CommitMsg.ViewNumber
		once := sync.GetOnly()
		once.Do(sync.Success)
	}

	s.lastVote = CommitMsg.ViewNumber

	//if ok, err := MatchingMsg(CommitMsg.MsgType, CommitMsg.ViewNumber, pb2.MsgType_COMMIT, *sync.ViewNumber()); !ok {
	//	return nil, err
	//}

	s.LockedQC = CommitMsg.Qc //更新LockedQC

	sync.TimerReset()
	s.SetState(Commit)

	block := CommitMsg.Block
	if chain.GetBlock(block.Hash) == nil {
		chain.Store(block) // 存储这一轮的区块
	}

	Sig, err := cryp.Sign(pb2.MsgType_COMMIT_VOTE, *sync.ViewNumber(), CommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	CommitVoteMsg := &pb2.VoteRequest{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  Sig,
		Hash:       CommitMsg.Hash,
		MsgType:    pb2.MsgType_COMMIT_VOTE,
	}
	var leader pb2.HotstuffClient
	leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]

	//模拟投票处理和传输时延
	time.Sleep(d.GetLatency())

	if !d.DenyVote {
		go leader.VoteCommit(context.Background(), CommitVoteMsg)
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Decide(ctx context.Context, DecideMsg *pb2.DecideMsg) (*emptypb.Empty, error) {
	s.WaitForState(Commit)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", DecideMsg.Id, " 的Decide消息")

	//确保消息来自主节点
	if DecideMsg.Id != sync.GetLeader(DecideMsg.ViewNumber) {
		log.Println("Decide消息的提议者不是当前视图的领导者")
		return nil, fmt.Errorf("DecideMsg is not valid")
	}

	//检查提案是否满足接收条件
	if !s.VoteRuleDecide(DecideMsg) {
		log.Println("Decide消息不安全")
		return nil, fmt.Errorf("DecideMsg is not valid")
	}

	//检查区块的视图号/高度是否满足同一高度仅能投票一次
	if DecideMsg.ViewNumber < s.lastVote {
		log.Println("DecideMsg 消息的视图号小于上一次投票的视图号，拒绝投票")
		return nil, fmt.Errorf("DecideMsg Msg is too old")
	}

	// 切换到合适视图
	if *sync.ViewNumber() < DecideMsg.ViewNumber {
		sync.ViewNumberSet(DecideMsg.ViewNumber)
		s.TempViewNumber = DecideMsg.ViewNumber
		once := sync.GetOnly()
		once.Do(sync.Success)
	}

	s.lastVote = DecideMsg.ViewNumber

	//if ok, err := MatchingMsg(DecideMsg.MsgType, DecideMsg.ViewNumber, pb2.MsgType_DECIDE, *sync.ViewNumber()); !ok {
	//	return nil, err
	//}
	sync.TimerReset()
	s.SetState(Switching)

	block := DecideMsg.Block
	if chain.GetBlock(block.Hash) == nil {
		chain.Store(block) // 存储这一轮的区块
	}

	chain.WriteToFile(block)

	sig, err := cryp.Sign(pb2.MsgType_NEW_VIEW, *sync.ViewNumber()+1, DecideMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}

	NewViewMsg := &pb2.NewViewMsg{
		ViewNumber: *sync.ViewNumber() + 1,
		Voter:      s.ID,
		Signature:  sig,
		Hash:       DecideMsg.Hash,
		MsgType:    pb2.MsgType_NEW_VIEW,
		Qc:         s.PrepareQC,
	}

	var leader pb2.HotstuffClient
	leader = *modules.MODULES.ReplicaClient[sync.GetLeader(*sync.ViewNumber()+1)]

	//模拟投票处理和传输时延
	time.Sleep(d.GetLatency())

	if !d.DenyVote {
		go leader.NewView(context.Background(), NewViewMsg)
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) NextView() { //所有的wait for阶段超时都会调用这个函数
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)
	for {
		select {
		case <-sync.Timeout():
			s.SetState(Switching)
			s.TempViewNumber++ //正常情况下TempViewNumber应该等于ViewNumber
			s.TimeoutRecord++
			s.RecordTimeOutLogToFile()
			var QC = s.PrepareQC

			sig, err := cryp.Sign(pb2.MsgType_NEW_VIEW, s.TempViewNumber, s.PrepareQC.BlockHash)
			if err != nil {
				log.Println("部分签名失败")
			}
			var leader pb2.HotstuffClient
			leader = *modules.MODULES.ReplicaClient[sync.GetLeader(s.TempViewNumber)]
			NewViewMsg := &pb2.NewViewMsg{
				// ProposalId: nil,
				ViewNumber: s.TempViewNumber,
				Voter:      s.ID,
				Signature:  sig,
				Hash:       s.PrepareQC.BlockHash,
				MsgType:    pb2.MsgType_NEW_VIEW,
				Qc:         QC,
			}
			log.Println("向节点 ", sync.GetLeader(s.TempViewNumber), " 发送超时临时视图 ", s.TempViewNumber, " 的NewView消息，当前视图: ", *sync.ViewNumber())
			leader.NewView(context.Background(), NewViewMsg)
		case <-s.ctx.Done():
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
