package hotstuff

import (
	"context"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"fmt"
	"log"
	"net"
	stsync "sync"

	"go.dedis.ch/kyber/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	sync  = modules.MODULES.Synchronizer
	cryp  = modules.MODULES.SignerAndVerifier
	chain = modules.MODULES.Chain
)

type ReplicaServer struct {
	stsync.Mutex
	sigs      map[kyber.Point][]byte
	count     int
	threshold int
	wg        stsync.WaitGroup
	once      stsync.Once

	ID        int32
	PrepareQC *pb.QC
	LockedQC  *pb.QC

	pb.UnimplementedHotstuffServer
}

func (s *ReplicaServer) NewView(ctx context.Context, NewViewMsg *pb.NewViewMsg) (*emptypb.Empty, error) {
	if !MatchingMsg(NewViewMsg.MsgType, NewViewMsg.ViewNumber, pb.MsgType_NEW_VIEW, curViewNumber) {
		return nil, fmt.Errorf("newview msg type is not NEW_VIEW")
	}
	//todo 此处应做签名校验，暂时省略
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的Propose
			//todo 获取QC
			ProposalMsg := &pb.Proposal{
				Block: nil, //暂未获取
				Qc:    nil, //暂未获取
				// Aggqc: nil, //hotstuff中用不到
				// ProposalId: 0, //暂未获取
				Proposer:   s.ID,
				ViewNumber: curViewNumber,
				Signature:  nil, //暂未获取
				// Timestamp:  0, //暂时用不到
				MsgType: pb.MsgType_PREPARE,
			}
			for _, client := range modules.MODULES.ReplicaClient {
				(*client).Propose(sync.GetContext(), ProposalMsg)
			}
			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器
		})
	}
	return &emptypb.Empty{}, nil
}

// 作为副本接收到主节点的提案后进行的处理
func (s *ReplicaServer) Propose(ctx context.Context, Proposal *pb.Proposal) (*emptypb.Empty, error) {
	if !MatchingMsg(Proposal.MsgType, Proposal.ViewNumber, pb.MsgType_PREPARE, curViewNumber) {
		return nil, fmt.Errorf("proposal msg type is not PREPARE")
	}
	if !SafeNode(Proposal.Block, Proposal.Qc) {
		return nil, fmt.Errorf("proposal is not safe")
	}
	//todo: 重置计时器
	sync.TimerReset()

	// 临时存储区块
	chain.StoreToTemp(Proposal.Block)

	sig, err := cryp.PartSign(Proposal.Signature)
	if err != nil {
		log.Println("部分签名失败")
	}
	PrepareVoteMsg := &pb.VoteRequest{
		ProposalId: Proposal.ProposalId,
		ViewNumber: curViewNumber,
		Voter:      s.ID,
		MsgType:    pb.MsgType_PREPARE_VOTE,
		Signature:  sig,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	leader.VotePrepare(sync.GetContext(), PrepareVoteMsg)
	return &emptypb.Empty{}, nil
}

// 作为主节点接收副本对准备消息的投票
func (s *ReplicaServer) VotePrepare(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {
	if !MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_PREPARE_VOTE, curViewNumber) {
		return nil, fmt.Errorf("投票消息类型不匹配")
	}
	//todo: 此处应做签名校验，暂时省略
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的PreCommit
			//todo 获取QC

			PreCommitMsg := &pb.Precommit{
				Id:         s.ID,
				MsgType:    pb.MsgType_PRE_COMMIT,
				ViewNumber: curViewNumber,
				Qc:         nil, //暂未获取
				Hash:       vote.Hash,
				Signature:  nil, //暂未获取
			}
			for _, client := range modules.MODULES.ReplicaClient {
				(*client).PreCommit(sync.GetContext(), PreCommitMsg)
			}

			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器

		})
	}
	return &emptypb.Empty{}, nil
}

// 作为副本接收到主节点的PreCommit消息后的处理
func (s *ReplicaServer) PreCommit(ctx context.Context, PrecommitMsg *pb.Precommit) (*emptypb.Empty, error) {
	if !MatchingMsg(PrecommitMsg.MsgType, PrecommitMsg.ViewNumber, pb.MsgType_PRE_COMMIT, curViewNumber) {
		return nil, fmt.Errorf("precommit msg type is not PRE_COMMIT")
	}

	sync.TimerReset()
	chain.Store(chain.GetBlockFromTemp(PrecommitMsg.Hash)) // 存储这一轮的区块
	sig, err := cryp.PartSign(PrecommitMsg.Signature)
	if err != nil {
		log.Println("部分签名失败")
	}
	PreCommitVoteMsg := &pb.VoteRequest{
		ViewNumber: curViewNumber,
		Voter:      s.ID,
		Signature:  sig,
		Hash:       PrecommitMsg.Hash,
		MsgType:    pb.MsgType_PRE_COMMIT_VOTE,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	leader.VotePreCommit(sync.GetContext(), PreCommitVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) VotePreCommit(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {
	if !MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_PRE_COMMIT_VOTE, curViewNumber) {
		return nil, fmt.Errorf("投票消息类型不匹配")
	}
	//todo 此处应做签名校验，暂时省略
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的Commit
			//todo 获取QC
			CommitMsg := &pb.CommitMsg{
				Id:         s.ID,
				MsgType:    pb.MsgType_COMMIT,
				ViewNumber: curViewNumber,
				Qc:         nil, //暂未获取
				Hash:       vote.Hash,
				Signature:  nil, //暂未获取
			}
			for _, client := range modules.MODULES.ReplicaClient {
				(*client).Commit(sync.GetContext(), CommitMsg)
			}

			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器
		})
	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Commit(ctx context.Context, CommitMsg *pb.CommitMsg) (*emptypb.Empty, error) {
	if !MatchingMsg(CommitMsg.MsgType, CommitMsg.ViewNumber, pb.MsgType_COMMIT, curViewNumber) {
		return nil, fmt.Errorf("commit msg type is not COMMIT")
	}
	sync.TimerReset() //重置计时器
	Sig, err := cryp.PartSign(CommitMsg.Signature)
	if err != nil {
		log.Println("部分签名失败")
	}
	CommitVoteMsg := &pb.VoteRequest{
		ViewNumber: curViewNumber,
		Voter:      s.ID,
		Signature:  Sig,
		Hash:       CommitMsg.Hash,
		MsgType:    pb.MsgType_COMMIT_VOTE,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	leader.VoteCommit(sync.GetContext(), CommitVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) VoteCommit(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {
	if !MatchingMsg(vote.MsgType, vote.ViewNumber, pb.MsgType_COMMIT_VOTE, curViewNumber) {
		return nil, fmt.Errorf("投票消息类型不匹配")
	}
	//todo 此处应做签名校验，暂时省略
	s.Lock()
	s.count++
	count := s.count
	s.Unlock()
	if count >= s.threshold { //条件达成，开始执行下一阶段
		sync.TimerReset()  //重置计时器
		s.once.Do(func() { //调用其他副本的Decide
			//todo 获取QC
			DecideMsg := &pb.DecideMsg{
				Id:         s.ID,
				MsgType:    pb.MsgType_DECIDE,
				ViewNumber: curViewNumber,
				Qc:         nil, //暂未获取
				Hash:       vote.Hash,
				Signature:  nil, //暂未获取
			}
			for _, client := range modules.MODULES.ReplicaClient {
				(*client).Decide(sync.GetContext(), DecideMsg)
			}
			s.once = stsync.Once{} //创建新的once实例
			s.count = 0            //重置计数器
		})

	}
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Decide(ctx context.Context, DecideMsg *pb.DecideMsg) (*emptypb.Empty, error) {
	if !MatchingMsg(DecideMsg.MsgType, DecideMsg.ViewNumber, pb.MsgType_DECIDE, curViewNumber) {
		return nil, fmt.Errorf("decide msg type is not DECIDE")
	}
	sync.TimerReset() //重置计时器
	sig, err := cryp.PartSign(DecideMsg.Signature)
	if err != nil {
		log.Println("部分签名失败")
	}
	curViewNumber++
	NewViewMsg := &pb.NewViewMsg{
		Id:         s.ID,
		MsgType:    pb.MsgType_NEW_VIEW,
		ViewNumber: curViewNumber,
		Qc:         s.PrepareQC,
		Signature:  sig,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()+1]
	leader.NewView(sync.GetContext(), NewViewMsg)
	return &emptypb.Empty{}, nil
}

func NewReplicaServer() *grpc.Server {
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
		log.Println("副本客户端连s接失败:", err)
		return nil
	}
	defer conn.Close()
	log.Println("副本客户端连接成功")
	client := pb.NewHotstuffClient(conn)
	modules.MODULES.ReplicaClient[id] = &client
	// modules.MODULES.ReplicaPubKey[id] = cryp.GetPubKey()
	return &client
}

//todo 超时之后重置：once 计时器
