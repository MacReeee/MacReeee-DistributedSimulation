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
	sync.StoreVote(pb2.MsgType_PREPARE_VOTE, vote)
	s.mu.Lock()
	s.count++
	count := s.count
	once := sync.GetOnce(pb2.MsgType_PREPARE_VOTE)
	s.mu.Unlock()
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
			time.Sleep(d.GetProcessTime())

			for _, client := range modules.MODULES.ReplicaClient {
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

	//如果已经触发阈值条件，不在接收后续的PreCommit投票
	//_, _, once := sync.GetVoter(pb.MsgType_PRE_COMMIT_VOTE)
	once := sync.GetOnce(pb2.MsgType_PRE_COMMIT_VOTE)
	if once.IsDone() {
		return nil, nil
	}

	log.Println("视图 ", *sync.ViewNumber(), ":接收到来自 ", vote.Voter, " 的PreCommit投票")
	if ok, err := MatchingMsg(vote.MsgType, vote.ViewNumber, pb2.MsgType_PRE_COMMIT_VOTE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		log.Println("Pre_Commit Vote 的签名验证失败")
		return nil, fmt.Errorf("Pre_Commit Vote 的签名验证失败")
	}
	sync.StoreVote(pb2.MsgType_PRE_COMMIT_VOTE, vote)
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()
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
			time.Sleep(d.GetProcessTime())

			for _, client := range modules.MODULES.ReplicaClient {
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

	//如果已经触发阈值条件，不在接收后续的Commit投票
	//_, _, once := sync.GetVoter(pb.MsgType_COMMIT_VOTE)
	once := sync.GetOnce(pb2.MsgType_COMMIT_VOTE)
	if once.IsDone() {
		return nil, nil
	}

	if ok, err := MatchingMsg(vote.MsgType, vote.ViewNumber, pb2.MsgType_COMMIT_VOTE, *sync.ViewNumber()); !ok {
		return nil, err
	}
	// 签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", vote.MsgType, vote.ViewNumber, vote.Hash))
	if !cryp.Verify(vote.Voter, msg, vote.Signature) {
		return nil, fmt.Errorf("commit vote signature is not valid")
	}
	sync.StoreVote(pb2.MsgType_COMMIT_VOTE, vote)
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()
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
			time.Sleep(d.GetProcessTime())

			for _, client := range modules.MODULES.ReplicaClient {
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

	//如果已经触发阈值条件，不在接收后续的NewView消息
	once := sync.GetOnce(pb2.MsgType_NEW_VIEW)
	if once.IsDone() {
		return nil, nil
	}

	if ok, err := MatchingMsg(NewViewMsg.MsgType, NewViewMsg.ViewNumber, pb2.MsgType_NEW_VIEW, *sync.ViewNumber()); !ok {
		log.Println("消息类型不匹配")
		return nil, err
	}
	//签名校验
	msg := []byte(fmt.Sprintf("%d,%d,%x", NewViewMsg.MsgType, NewViewMsg.ViewNumber, NewViewMsg.Hash))
	if !cryp.Verify(NewViewMsg.Voter, msg, NewViewMsg.Signature) {
		log.Println("NewView消息签名验证失败")
		return nil, fmt.Errorf("newview msg signature is not valid")
	}
	sync.StoreVote(pb2.MsgType_NEW_VIEW, nil, NewViewMsg)
	s.mu.Lock()
	s.count++
	count := s.count
	s.mu.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset() //重置计时器
		var ProposalMsg = &pb2.Proposal{}
		once.Do(func() { //调用其他副本的Propose
			// 获取QC
			HighQC := sync.HighQC()
			// 对QC签名作为自己的签名
			qcjson := QCMarshal(HighQC)
			sig, _ := cryp.NormSign(qcjson)
			// 创建区块
			block := chain.CreateBlock(HighQC.BlockHash, *sync.ViewNumber()+1, HighQC, []byte("CMD of View: "+strconv.Itoa(int(*sync.ViewNumber()+1))), s.ID)
			if HighQC.ViewNumber+1 < *sync.ViewNumber() {
				log.Println("高QC的视图号小于当前视图号")
			}
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

			time.Sleep(10 * time.Millisecond)

			log.Println("尝试发送视图 ", *sync.ViewNumber()+1, " 的提案 ")

			// 视图成功并退出，视图成功理应在接收到Prepare消息时触发
			// 放在此处是防止其他节点投票到来时主节点还未切换视图
			// 对于主节点来说放在这里和放在Prepare函数中等效
			//ViewSuccess(sync)
			// 同理，对于主节点来说存储临时区块的操作应该放在Prepare函数中
			// 但是为了保证视图对齐，这里也存储临时区块
			chain.StoreToTemp(block)

			for _, client := range modules.MODULES.ReplicaClient {
				go (*client).Prepare(context.Background(), ProposalMsg)
			}

			s.count = 0 //重置计数器
		})
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) NextView() { //所有的wait for阶段超时都会调用这个函数，//todo 记得go调用
	var (
		sync = modules.MODULES.Synchronizer
		cryp = modules.MODULES.Signer
		// chain = modules.MODULES.Chain
	)
	for {
		select {
		case <-sync.Timeout():
			s.SetState(Switching)
			ViewNow := *sync.ViewNumber()
			ViewNow++
			//ViewSuccess(sync) //退出视图，避免同一视图一直超时无法进入下一视图
			var QC = s.PrepareQC
			//对QC签名作为自己的签名
			sig, err := cryp.Sign(pb2.MsgType_NEW_VIEW, *sync.ViewNumber(), s.PrepareQC.BlockHash)
			if err != nil {
				log.Println("部分签名失败")
			}
			var leader pb2.HotstuffClient
			leader = *modules.MODULES.ReplicaClient[sync.GetLeader(ViewNow)]
			NewViewMsg := &pb2.NewViewMsg{
				// ProposalId: nil,
				ViewNumber: ViewNow,
				Voter:      s.ID,
				Signature:  sig,
				Hash:       s.PrepareQC.BlockHash,
				MsgType:    pb2.MsgType_NEW_VIEW,
				Qc:         QC,
			}
			leader.NewView(context.Background(), NewViewMsg)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
