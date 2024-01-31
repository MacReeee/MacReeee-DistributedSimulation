package hotstuff

import (
	"context"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"fmt"
	"log"
	"net"
	stsync "sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	sync    = modules.MODULES.Synchronizer
	crypto_ = modules.MODULES.SignerAndVerifier
	chain   = modules.MODULES.Chain
)

type ReplicaServer struct {
	stsync.Mutex
	count     int
	threshold int
	wg        stsync.WaitGroup

	pb.UnimplementedHotstuffServer
}

// 作为副本接收到主节点的提案后进行的处理
func (*ReplicaServer) Propose(ctx context.Context, Proposal *pb.Proposal) (*emptypb.Empty, error) {
	if !MatchingMsg(Proposal, pb.MsgType_PREPARE, curViewNumber) {
		return nil, fmt.Errorf("proposal msg type is not PREPARE")
	}
	if !SafeNode(Proposal.Block, Proposal.Qc) {
		return nil, fmt.Errorf("proposal is not safe")
	}
	//todo: 重置计时器
	sync.TimerReset()

	// 临时存储区块
	chain.StoreTemp(Proposal.Block)

	sig, err := crypto_.PartSign(Proposal.Signature)
	if err != nil {
		log.Println("部分签名失败")
	}
	PrepareVoteMsg := &pb.VoteRequest{
		ProposalId: Proposal.ProposalId,
		ViewNumber: curViewNumber,
		Voter:      ReplicaID,
		MsgType:    pb.MsgType_PREPARE_VOTE,
		Signature:  sig,
	}
	leader := *modules.MODULES.ReplicaClient[sync.GetLeader()]
	leader.Vote(sync.GetContext(), PrepareVoteMsg)
	return &emptypb.Empty{}, nil
}

func (s *ReplicaServer) Vote(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {

	s.Lock()
	s.count++
	if s.count == 1 {
		s.wg.Add(s.threshold) // 第一个投票到达时设置等待组的计数器
	}
	if s.count >= s.threshold { // 达到阈值，让所有客户端继续执行
		s.wg.Done()
		s.count = 0 // 重置计数器，等待下一批客户端调用
	}
	s.Unlock()

	s.wg.Wait() // 如果还没达到阈值，此处等待
	return nil, nil
}
func (*ReplicaServer) PreCommit(ctx context.Context, PrecommitMsg *pb.Precommit) (*emptypb.Empty, error) {

	return nil, nil
}
func (*ReplicaServer) Commit(ctx context.Context, CommitMsg *pb.CommitMsg) (*emptypb.Empty, error) {

	return nil, nil
}
func (*ReplicaServer) Decide(ctx context.Context, DecideMsg *pb.DecideMsg) (*emptypb.Empty, error) {

	return nil, nil
}
func (*ReplicaServer) NewView(ctx context.Context, NewViewMsg *pb.NewViewMsg) (*emptypb.Empty, error) {

	return nil, nil
}

func NewReplicaServer() *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ReplicaID+4000))
	if err != nil {
		log.Println("副本服务监听失败:", err)
	}
	server := grpc.NewServer()
	pb.RegisterHotstuffServer(server, &ReplicaServer{
		threshold: 2,
		count:     0,
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
	}
	defer conn.Close()
	log.Println("副本客户端连接成功")
	client := pb.NewHotstuffClient(conn)
	modules.MODULES.ReplicaClient[id] = &client
	return &client
}
