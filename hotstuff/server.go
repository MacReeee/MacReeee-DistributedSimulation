package hotstuff

import (
	"context"
	"distributed/hotstuff/pb"
	"fmt"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type ReplicaServer struct {
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
	TempBlockMap[string(Proposal.Block.Hash)] = &Block{
		block:    Proposal.Block,
		proposer: Proposal.Proposer,
		children: nil,
	}

	//todo: 视图
	PrepareVoteMsg := &pb.VoteRequest{
		ProposalId: Proposal.ProposalId,
		ViewNumber: curViewNumber,
		Voter:      ReplicaID,
	}
	return &emptypb.Empty{}, nil
}

func (*ReplicaServer) Vote(context.Context, *pb.VoteRequest) (*emptypb.Empty, error) {

	return nil, nil
}
func (*ReplicaServer) PreCommit(context.Context, *pb.Precommit) (*emptypb.Empty, error) {

	return nil, nil
}
func (*ReplicaServer) Commit(context.Context, *pb.CommitMsg) (*emptypb.Empty, error) {

	return nil, nil
}
func (*ReplicaServer) Decide(context.Context, *pb.DecideMsg) (*emptypb.Empty, error) {

	return nil, nil
}
func (*ReplicaServer) NewView(context.Context, *pb.NewViewMsg) (*emptypb.Empty, error) {

	return nil, nil
}
