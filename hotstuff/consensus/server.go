package hotstuff

import (
	"context"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"fmt"
	"log"
	"net"
	"strconv"
	stsync "sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	sync = modules.MODULES.Synchronizer
	// cryp  = modules.MODULES.SignerAndVerifier
	cryp  = modules.MODULES.Signer
	chain = modules.MODULES.Chain
)

type ReplicaServer struct {
	stsync.Mutex
	// sigs      map[kyber.Point][]byte
	count     int
	threshold int
	wg        stsync.WaitGroup
	once      stsync.Once

	ID        int32
	PrepareQC *pb.QC
	LockedQC  *pb.QC

	pb.UnimplementedHotstuffServer
}

// 这是收到NewView消息后的处理，开启新视图
func (s *ReplicaServer) NewView(ctx context.Context, NewViewMsg *pb.NewViewMsg) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(NewViewMsg.MsgType, NewViewMsg.ViewNumber, pb.MsgType_NEW_VIEW, viewNum) {
		return nil, fmt.Errorf("newview msg type is not NEW_VIEW")
	}
	//签名校验
	if !cryp.Verify(NewViewMsg.Id, NewViewMsg.Signature, QCMarshal(NewViewMsg.Qc)) {
		return nil, fmt.Errorf("newview msg signature is not valid")
	}
	sync.StoreVote(pb.MsgType_NEW_VIEW, nil, NewViewMsg)
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的Propose
			// 获取QC
			HighQC := sync.HighQC()
			// 对QC签名作为自己的签名
			qcjson := QCMarshal(HighQC)
			sig, _ := cryp.NormSign(qcjson)
			// 创建区块
			block := chain.CreateBlock(NewViewMsg.Qc.BlockHash, sync.ViewNumber(), NewViewMsg.Qc, []byte("CMD of View:  "+strconv.Itoa(int(sync.ViewNumber()))), s.ID)

			ProposalMsg := &pb.Proposal{
				Block: block,
				Qc:    HighQC,
				// Aggqc: nil, //hotstuff中用不到
				// ProposalId: 0, //暂未获取，未考虑清楚是否需要该字段
				Proposer:   s.ID,
				ViewNumber: viewNum,
				Signature:  sig,
				// Timestamp:  0, //暂时用不到
				MsgType: pb.MsgType_PREPARE,
			}

			for _, client := range modules.MODULES.ReplicaClient {
				// (*client).Propose(sync.GetContext(), ProposalMsg)
				(*client).Propose(context.Background(), ProposalMsg)
			}

			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器
			// s.Vote.NewView = []*pb.NewViewMsg{} //清空NewView
		})
	}
	return &emptypb.Empty{}, nil
}

// 作为副本接收到主节点的提案后进行的处理
func (s *ReplicaServer) Propose(ctx context.Context, Proposal *pb.Proposal) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(Proposal.MsgType, Proposal.ViewNumber, pb.MsgType_PREPARE, viewNum) {
		return nil, fmt.Errorf("proposal msg type is not PREPARE")
	}
	if !SafeNode(Proposal.Block, Proposal.Qc) {
		return nil, fmt.Errorf("proposal is not safe")
	}
	//todo: 重置计时器
	sync.TimerReset()

	// 临时存储区块
	chain.StoreToTemp(Proposal.Block)

	//对提案进行签名
	sig, err := cryp.Sign(pb.MsgType_PREPARE, viewNum, Proposal.Block.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PrepareVoteMsg := &pb.VoteRequest{
		// ProposalId: Proposal.ProposalId,
		ViewNumber: viewNum,
		Voter:      s.ID,
		Signature:  sig,
		Hash:       Proposal.Block.Hash,
		MsgType:    pb.MsgType_PREPARE_VOTE,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	// leader.VotePrepare(sync.GetContext(), PrepareVoteMsg)
	leader.VotePrepare(context.Background(), PrepareVoteMsg)
	return &emptypb.Empty{}, nil
}

// 作为主节点接收副本对准备消息的投票
func (s *ReplicaServer) VotePrepare(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_PREPARE_VOTE, viewNum) {
		return nil, fmt.Errorf("投票消息类型不匹配")
	}
	// 签名校验
	if !cryp.Verify(vote.Voter, vote.Signature, vote.Hash) {
		return nil, fmt.Errorf("prepare vote signature is not valid")
	}
	sync.StoreVote(pb.MsgType_PREPARE_VOTE, vote)
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的PreCommit
			// 合成QC
			voters, sigs := sync.GetVoter(pb.MsgType_PREPARE_VOTE)
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb.MsgType_PREPARE_VOTE,
				ViewNumber:   sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)

			PreCommitMsg := &pb.Precommit{
				Id:         s.ID,
				MsgType:    pb.MsgType_PRE_COMMIT,
				ViewNumber: sync.ViewNumber(),
				Qc:         QC,
				Hash:       vote.Hash,
				Signature:  sig,
			}
			for _, client := range modules.MODULES.ReplicaClient {
				// (*client).PreCommit(sync.GetContext(), PreCommitMsg)
				(*client).PreCommit(context.Background(), PreCommitMsg)
			}

			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器

		})
	}
	return &emptypb.Empty{}, nil
}

// 作为副本接收到主节点的PreCommit消息后的处理
func (s *ReplicaServer) PreCommit(ctx context.Context, PrecommitMsg *pb.Precommit) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(PrecommitMsg.MsgType, PrecommitMsg.ViewNumber, pb.MsgType_PRE_COMMIT, viewNum) {
		return nil, fmt.Errorf("precommit msg type is not PRE_COMMIT")
	}

	s.PrepareQC = PrecommitMsg.Qc //更新PrepareQC

	sync.TimerReset()
	chain.Store(chain.GetBlockFromTemp(PrecommitMsg.Hash)) // 存储这一轮的区块

	sig, err := cryp.Sign(pb.MsgType_PRE_COMMIT, viewNum, PrecommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	PreCommitVoteMsg := &pb.VoteRequest{
		ViewNumber: viewNum,
		Voter:      s.ID,
		Signature:  sig,
		Hash:       PrecommitMsg.Hash,
		MsgType:    pb.MsgType_PRE_COMMIT_VOTE,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	// leader.VotePreCommit(sync.GetContext(), PreCommitVoteMsg)
	leader.VotePreCommit(context.Background(), PreCommitVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) VotePreCommit(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_PRE_COMMIT_VOTE, viewNum) {
		return nil, fmt.Errorf("投票消息类型不匹配")
	}
	// 签名校验
	if !cryp.Verify(vote.Voter, vote.Signature, vote.Hash) {
		return nil, fmt.Errorf("precommit vote signature is not valid")
	}
	sync.StoreVote(pb.MsgType_PRE_COMMIT_VOTE, vote)
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的Commit
			// 合成QC
			voters, sigs := sync.GetVoter(pb.MsgType_PRE_COMMIT_VOTE)
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb.MsgType_PRE_COMMIT_VOTE,
				ViewNumber:   sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			CommitMsg := &pb.CommitMsg{
				Id:         s.ID,
				MsgType:    pb.MsgType_COMMIT,
				ViewNumber: viewNum,
				Qc:         QC,
				Hash:       vote.Hash,
				Signature:  sig,
			}
			for _, client := range modules.MODULES.ReplicaClient {
				// (*client).Commit(sync.GetContext(), CommitMsg)
				(*client).Commit(context.Background(), CommitMsg)
			}

			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器
		})
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Commit(ctx context.Context, CommitMsg *pb.CommitMsg) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(CommitMsg.MsgType, CommitMsg.ViewNumber, pb.MsgType_COMMIT, viewNum) {
		return nil, fmt.Errorf("commit msg type is not COMMIT")
	}

	s.LockedQC = CommitMsg.Qc //更新LockedQC

	sync.TimerReset() //重置计时器

	Sig, err := cryp.Sign(pb.MsgType_COMMIT, viewNum, CommitMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	CommitVoteMsg := &pb.VoteRequest{
		ViewNumber: viewNum,
		Voter:      s.ID,
		Signature:  Sig,
		Hash:       CommitMsg.Hash,
		MsgType:    pb.MsgType_COMMIT_VOTE,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	// leader.VoteCommit(sync.GetContext(), CommitVoteMsg)
	leader.VoteCommit(context.Background(), CommitVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) VoteCommit(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_COMMIT_VOTE, viewNum) {
		return nil, fmt.Errorf("投票消息类型不匹配")
	}
	// 签名校验
	if !cryp.Verify(vote.Voter, vote.Signature, vote.Hash) {
		return nil, fmt.Errorf("commit vote signature is not valid")
	}
	sync.StoreVote(pb.MsgType_COMMIT_VOTE, vote)
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的Decide
			// 合成QC
			voters, sigs := sync.GetVoter(pb.MsgType_COMMIT_VOTE)
			qc, aggpub, _ := cryp.ThreshMock(voters, sigs)
			var QC = &pb.QC{
				BlsSignature: qc,
				AggPubKey:    aggpub,
				Voter:        voters,
				MsgType:      pb.MsgType_COMMIT_VOTE,
				ViewNumber:   sync.ViewNumber(),
				BlockHash:    vote.Hash,
			}

			// 对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			DecideMsg := &pb.DecideMsg{
				Id:         s.ID,
				MsgType:    pb.MsgType_DECIDE,
				ViewNumber: viewNum,
				Qc:         QC, //暂未获取
				Hash:       vote.Hash,
				Signature:  sig, //暂未获取
			}
			for _, client := range modules.MODULES.ReplicaClient {
				// (*client).Decide(sync.GetContext(), DecideMsg)
				(*client).Decide(context.Background(), DecideMsg)
			}
			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器
		})

	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Decide(ctx context.Context, DecideMsg *pb.DecideMsg) (*emptypb.Empty, error) {
	viewNum := sync.ViewNumber()
	if !MatchingMsg(DecideMsg.MsgType, DecideMsg.ViewNumber, pb.MsgType_DECIDE, viewNum) {
		return nil, fmt.Errorf("decide msg type is not DECIDE")
	}
	sync.TimerReset() //重置计时器
	sig, err := cryp.Sign(pb.MsgType_NEW_VIEW, viewNum, DecideMsg.Hash)
	if err != nil {
		log.Println("部分签名失败")
	}
	curViewNumber++
	newestBlock := chain.GetBlock(DecideMsg.Hash)
	//剪枝，并存储已经稳定上链的区块
	chain.PruneBlock(chain.GetBlock(newestBlock.ParentHash), newestBlock)
	//视图成功并退出
	ViewSuccess(sync)
	NewViewMsg := &pb.NewViewMsg{
		Id:         s.ID,
		MsgType:    pb.MsgType_NEW_VIEW,
		ViewNumber: sync.ViewNumber(),
		Qc:         s.PrepareQC,
		Signature:  sig,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	// leader.NewView(sync.GetContext(), NewViewMsg)
	leader.NewView(context.Background(), NewViewMsg)
	return &emptypb.Empty{}, nil
}

func NextView(s *ReplicaServer) { //所有的wait for阶段如果超时，都会调用这个函数，//todo 记得go调用
	for {
		select {
		case <-sync.Timeout():
			var QC = s.PrepareQC
			//对QC签名作为自己的签名
			qcjson := QCMarshal(QC)
			sig, _ := cryp.NormSign(qcjson)
			leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
			NewViewMsg := &pb.NewViewMsg{
				Id:         s.ID,
				MsgType:    pb.MsgType_NEW_VIEW,
				ViewNumber: sync.ViewNumber(),
				Qc:         QC,
				Signature:  sig, //todo应当对QC进行签名，暂时省略
			}
			leader.NewView(context.Background(), NewViewMsg)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func NewReplicaServer(id int32) *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ReplicaID+4000))
	if err != nil {
		log.Println("副本服务监听失败:", err)
	}
	server := grpc.NewServer()
	pb.RegisterHotstuffServer(server, &ReplicaServer{
		threshold: 3,
		count:     0,
		wg:        stsync.WaitGroup{},
		once:      stsync.Once{},
	})
	log.Println("副本服务启动成功")
	server.Serve(listener)
	modules.MODULES.ReplicaServer = server
	return server
}

func NewReplicaClient(id int32) *pb.HotstuffClient {
	conn, err := grpc.Dial(fmt.Sprintf(":%d", id+4000), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("副本客户端连接失败:", err)
		return nil
	}
	defer conn.Close()
	log.Println("副本客户端连接成功")
	client := pb.NewHotstuffClient(conn)
	modules.MODULES.ReplicaClient[id] = &client
	// modules.MODULES.ReplicaPubKey[id] = cryp.GetPubKey()
	return &client
}

//如果连续失败多个视图，如何保障节点之间的视图对齐？
