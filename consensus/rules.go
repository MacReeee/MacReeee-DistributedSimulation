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
		log.Println("Prepare: 接收到的区块视图号小于当前视图号，拒绝投票")
		return false
	}

	parent := chain.GetBlock(proposal.Block.ParentHash)
	if parent == nil {
		log.Println("Prepare: 接收到的区块的父区块不存在，拒绝投票")
		return false
	}
	// Rule 2: 只给父视图大于等于锁定块视图的块投票
	if parent.ViewNumber < s.LockedQC.ViewNumber {
		log.Println("Prepare: 接收到的区块的父区块视图号小于当前锁定块视图号，拒绝投票")
		return false
	}

	return true
}

func (s *ReplicaServer) VoteRulePreCommit(PreCommit *pb.Precommit) bool {
	var (
		sync  = modules.MODULES.Synchronizer
		chain = modules.MODULES.Chain
	)
	// Rule 1: 只给增加的轮次投票
	if PreCommit.Block.ViewNumber < *sync.ViewNumber() {
		log.Println("PreCommit: 接收到的区块视图号小于当前视图号，拒绝投票")
		return false
	}

	parent := chain.GetBlock(PreCommit.Block.ParentHash)
	if parent == nil {
		log.Println("PreCommit: 接收到的区块的父区块不存在，拒绝投票")
		return false
	}
	// Rule 2: 只给父视图大于等于锁定块视图的块投票
	if parent.ViewNumber < s.LockedQC.ViewNumber {
		log.Println("PreCommit: 接收到的区块的父区块视图号小于当前锁定块视图号，拒绝投票")
		return false
	}
	return true
}

func (s *ReplicaServer) VoteRuleCommit(Commit *pb.CommitMsg) bool {
	var (
		sync  = modules.MODULES.Synchronizer
		chain = modules.MODULES.Chain
	)
	// Rule 1: 只给增加的轮次投票
	if Commit.Block.ViewNumber < *sync.ViewNumber() {
		log.Println("Commit: 接收到的区块视图号小于当前视图号，拒绝投票")
		return false
	}

	parent := chain.GetBlock(Commit.Block.ParentHash)
	if parent == nil {
		log.Println("Commit: 接收到的区块的父区块不存在，拒绝投票")
		return false
	}
	// Rule 2: 只给父视图大于等于锁定块视图的块投票
	if parent.ViewNumber < s.LockedQC.ViewNumber {
		log.Println("Commit: 接收到的区块的父区块视图号小于当前锁定块视图号，拒绝投票")
		return false
	}
	return true
}

func (s *ReplicaServer) VoteRuleDecide(Decide *pb.DecideMsg) bool {
	var (
		sync  = modules.MODULES.Synchronizer
		chain = modules.MODULES.Chain
	)
	// Rule 1: 只给增加的轮次投票
	if Decide.Block.ViewNumber < *sync.ViewNumber() {
		log.Println("Decide: 接收到的区块视图号小于当前视图号，拒绝投票")
		return false
	}

	parent := chain.GetBlock(Decide.Block.ParentHash)
	if parent == nil {
		log.Println("Decide: 接收到的区块的父区块不存在，拒绝投票")
		return false
	}
	// Rule 2: 只给父视图大于等于锁定块视图的块投票
	if parent.ViewNumber < s.LockedQC.ViewNumber-1 {
		log.Println("Decide: 接收到的区块的父区块视图号小于当前锁定块视图号，拒绝投票")
		return false
	}
	return true
}
