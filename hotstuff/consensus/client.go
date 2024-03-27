package hotstuff

import (
	"context"
	d "distributed/hotstuff/dependency"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"time"
)

func (s *ReplicaServer) Prepare(ctx context.Context, Proposal *pb.Proposal) (*emptypb.Empty, error) {
	// 设置状态，防止并发
	s.SetState(Prepare)
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	//fmt.Println("\n", "视图", *sync.ViewNumber(), "的log: ")
	fmt.Printf("\n")
	if sync.GetLeader(Proposal.ViewNumber) != s.ID {
		log.Println("接收到来自视图 ", *sync.ViewNumber()+1, "中节点 ", Proposal.Proposer, " 的提案，等待验证后进入新视图，当前视图: ", *sync.ViewNumber())
	} else {
		log.Println("接收到来自视图 ", *sync.ViewNumber(), "中节点 ", Proposal.Proposer, " 的提案，等待验证后进入新视图，当前视图: ", *sync.ViewNumber()-1)
	}

	//用以控制台控制中断
	if StopFlag {
		wg.Wait()
	}

	// 由于主节点在NewView阶段已经触发视图成功，所以此处针对主节点视图号和普通节点分别处理
	var vn int64
	if sync.GetLeader(Proposal.ViewNumber) == s.ID {
		vn = *sync.ViewNumber()
	} else {
		vn = *sync.ViewNumber() + 1
	}
	if ok, err := MatchingMsg(Proposal.MsgType, Proposal.ViewNumber, pb.MsgType_PREPARE, vn); !ok {
		//准备阶段尽管可能出现视图号不匹配，
		//但是如果视图号大于当前视图号，说明是新视图的提案，不应该被拒绝
		if Proposal.ViewNumber <= *sync.ViewNumber() {
			log.Println("消息类型不匹配")
			return nil, err
		} else if Proposal.ViewNumber > *sync.ViewNumber() {
			for Proposal.ViewNumber > *sync.ViewNumber() {
				ViewSuccess(sync)
			}
		}
	}
	if !s.SafeNode(Proposal.Block, Proposal.Qc) {
		log.Println("提案不安全")
		return nil, fmt.Errorf("proposal is not safe")
	}

	//视图成功并退出，如果是第一个视图，视图号此时从0变为1
	//如果不是主节点，则应该触发视图成功，主节点则在NewView中触发过
	if sync.GetLeader(Proposal.ViewNumber) != s.ID {
		ViewSuccess(sync)
	}

	// 临时存储区块
	if sync.GetLeader(Proposal.ViewNumber) != s.ID {
		chain.StoreToTemp(Proposal.Block)
	}

	//对提案进行签名
	sig, err := cryp.Sign(pb.MsgType_PREPARE_VOTE, *sync.ViewNumber(), Proposal.Block.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PrepareVoteMsg := &pb.VoteRequest{
		// ProposalId: Proposal.ProposalId,
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       Proposal.Block.Hash,
		MsgType:    pb.MsgType_PREPARE_VOTE,
	}

	var leader pb.HotstuffClient
	if d.DebugMode {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
		log.Println("视图 ", *sync.ViewNumber(), " 的领导者是: ", sync.GetLeader())
	}

	// 处理完成，状态置为Idle
	s.SetState(Idle)

	//模拟投票处理和传输时延
	time.Sleep(d.GetProcessTime())

	go leader.VotePrepare(context.Background(), PrepareVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) PreCommit(ctx context.Context, PrecommitMsg *pb.Precommit) (*emptypb.Empty, error) {
	s.waitForIdle()
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", PrecommitMsg.Id, " 的PreCommit消息")
	if ok, err := MatchingMsg(PrecommitMsg.MsgType, PrecommitMsg.ViewNumber, pb.MsgType_PRE_COMMIT, *sync.ViewNumber()); !ok {
		return nil, err
	}

	s.PrepareQC = PrecommitMsg.Qc //更新PrepareQC

	sync.TimerReset()
	block := chain.GetBlockFromTemp(PrecommitMsg.Hash)
	chain.Store(block) // 存储这一轮的区块

	sig, err := cryp.Sign(pb.MsgType_PRE_COMMIT_VOTE, *sync.ViewNumber(), PrecommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PreCommitVoteMsg := &pb.VoteRequest{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       PrecommitMsg.Hash,
		MsgType:    pb.MsgType_PRE_COMMIT_VOTE,
	}
	var leader pb.HotstuffClient
	if d.DebugMode {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
	}

	//模拟投票处理和传输时延
	time.Sleep(d.GetProcessTime())

	go leader.VotePreCommit(context.Background(), PreCommitVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Commit(ctx context.Context, CommitMsg *pb.CommitMsg) (*emptypb.Empty, error) {
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", CommitMsg.Id, " 的Commit消息")
	if ok, err := MatchingMsg(CommitMsg.MsgType, CommitMsg.ViewNumber, pb.MsgType_COMMIT, *sync.ViewNumber()); !ok {
		return nil, err
	}

	s.LockedQC = CommitMsg.Qc //更新LockedQC

	sync.TimerReset() //重置计时器

	Sig, err := cryp.Sign(pb.MsgType_COMMIT_VOTE, *sync.ViewNumber(), CommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	CommitVoteMsg := &pb.VoteRequest{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  Sig,
		Hash:       CommitMsg.Hash,
		MsgType:    pb.MsgType_COMMIT_VOTE,
	}
	var leader pb.HotstuffClient
	if d.DebugMode {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader()]
	}

	//模拟投票处理和传输时延
	time.Sleep(d.GetProcessTime())

	go leader.VoteCommit(context.Background(), CommitVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Decide(ctx context.Context, DecideMsg *pb.DecideMsg) (*emptypb.Empty, error) {
	var (
		sync  = modules.MODULES.Synchronizer
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", DecideMsg.Id, " 的Decide消息")
	if ok, err := MatchingMsg(DecideMsg.MsgType, DecideMsg.ViewNumber, pb.MsgType_DECIDE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	sync.TimerReset() //重置计时器
	sig, err := cryp.Sign(pb.MsgType_NEW_VIEW, *sync.ViewNumber(), DecideMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	curViewNumber++

	var newestBlock *pb.Block = chain.GetBlock(DecideMsg.Hash)
	if newestBlock != nil {
		//剪枝，并存储已经稳定上链的区块
		parent := chain.GetBlock(newestBlock.ParentHash)
		if parent != nil { //todo 无法定位parent可能不存在的问题，暂时用if判断 //已解决，问题出自newestBlock可能为空
			chain.PruneBlock(parent, newestBlock)
		}
		chain.WriteToFile(newestBlock)
	} else {
		// 依照原算法，区块不存在是正常现象，可能由于节点暂时掉线产生
		// 因此本来应该在此同步区块，但是不做区块同步对本次实验影响不大，因此不做处理
		log.Println("区块不存在")
		log.Println(newestBlock)
		writeFatalErr(fmt.Sprintf("节点 %d 的区块不存在: %x\n，当前视图: %d", s.ID, DecideMsg.Hash, *sync.ViewNumber()))
	}

	NewViewMsg := &pb.NewViewMsg{
		ViewNumber: *sync.ViewNumber(),
		Voter:      s.ID,
		Signature:  sig,
		Hash:       DecideMsg.Hash,
		MsgType:    pb.MsgType_NEW_VIEW,
		Qc:         s.PrepareQC,
	}

	var leader pb.HotstuffClient
	if d.DebugMode {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader(*sync.ViewNumber()+1)]
	} else {
		leader = *modules.MODULES.ReplicaClient[sync.GetLeader(*sync.ViewNumber()+1)]
	}

	//模拟投票处理和传输时延
	time.Sleep(d.GetProcessTime())

	go leader.NewView(context.Background(), NewViewMsg)
	return &emptypb.Empty{}, nil
}
