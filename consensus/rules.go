package hotstuff

import (
	"distributed/modules"
	"distributed/pb"
	"log"
)

func (s *ReplicaServer) VoteRule(proposal *pb.Proposal) bool {
	var (
		sync  = modules.MODULES.Synchronizer
		chain = modules.MODULES.Chain
	)
	// Rule 1: 只给增加的轮次投票
	if proposal.Block.ViewNumber < *sync.ViewNumber() {
		log.Println("接收到的区块视图号小于当前视图号，拒绝投票")
		return false
	}

	parent := chain.GetBlock(proposal.Block.ParentHash)
	if parent == nil {
		log.Println("接收到的区块的父区块不存在，拒绝投票")
		return false
	}
	// Rule 2: 只给父视图大于等于锁定块视图的块投票
	if parent.ViewNumber < s.LockedQC.ViewNumber {
		log.Println("接收到的区块的父区块视图号小于当前锁定块视图号，拒绝投票")
		return false
	}

	return true
}
