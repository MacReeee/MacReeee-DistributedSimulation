package hotstuff

import (
	"context"
	"distributed/hotstuff/blockchain"
	"distributed/hotstuff/pb"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"
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

	TempBlockMap[string(Proposal.Block.Hash)] = &blockchain.Block{
		Block:    Proposal.Block,
		Proposer: Proposal.Proposer,
		Children: nil,
	}

	//todo: 视图
	// PrepareVoteMsg := &pb.VoteRequest{
	// 	ProposalId: Proposal.ProposalId,
	// 	ViewNumber: curViewNumber,
	// 	Voter:      ReplicaID,
	// }

	return &emptypb.Empty{}, nil
}

func (*ReplicaServer) Vote(ctx context.Context, vote *pb.VoteRequest) (*emptypb.Empty, error) {

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
