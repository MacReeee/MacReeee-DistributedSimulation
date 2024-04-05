package view

import (
	"distributed/dependency"
	"distributed/modules"
	pb2 "distributed/pb"
	"log"
	"sync"
)

// 如果传入视图号，则返回该视图号对应的 Leader 编号，否则返回当前视图对应的 Leader 编号。
func (s *SYNC) GetLeader(viewnumber ...int64) int32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(viewnumber) == 0 {
		leader := int32(s.CurrentView) % dependency.NumReplicas
		if leader == 0 {
			return dependency.NumReplicas
		} else {
			return leader
		}
	}
	leader := int32(viewnumber[0]) % dependency.NumReplicas
	if leader == 0 {
		return dependency.NumReplicas
	} else {
		return leader
	}
}

func (s *SYNC) GetOnly() *sync.Once {
	return s.view.only
}

func (s *SYNC) ViewNumber() *int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.CurrentView
}

func (s *SYNC) Timeout() <-chan bool {
	return s.TimeoutChan
}

func (s *SYNC) StoreVote(msgType pb2.MsgType, NormalMsg *pb2.VoteRequest, NewViewMsg ...*pb2.NewViewMsg) int {
	s.view.mu.Lock()
	defer s.view.mu.Unlock()
	if NormalMsg != nil {
		switch msgType {
		case pb2.MsgType_PREPARE_VOTE:
			s.view.Vote.Prepare = append(s.view.Vote.Prepare, NormalMsg)
			s.view.Vote.PrepareVoter = append(s.view.Vote.PrepareVoter, NormalMsg.Voter)
			return len(s.view.Vote.Prepare)
		case pb2.MsgType_PRE_COMMIT_VOTE:
			s.view.Vote.PreCommit = append(s.view.Vote.PreCommit, NormalMsg)
			s.view.Vote.PreCommitVoter = append(s.view.Vote.PreCommitVoter, NormalMsg.Voter)
			return len(s.view.Vote.PreCommit)
		case pb2.MsgType_COMMIT_VOTE:
			s.view.Vote.Commit = append(s.view.Vote.Commit, NormalMsg)
			s.view.Vote.CommitVoter = append(s.view.Vote.CommitVoter, NormalMsg.Voter)
			return len(s.view.Vote.Commit)
		}
	}
	if NewViewMsg != nil {
		s.view.Vote.NewView = append(s.view.Vote.NewView, NewViewMsg...)
		s.view.Vote.NewViewVoter = append(s.view.Vote.NewViewVoter, NewViewMsg[0].ProposalId)
		return len(s.view.Vote.NewView)
	}
	return 0
}

func (s *SYNC) GetVoter(msgType pb2.MsgType) ([]int32, [][]byte, *dependency.OnceWithDone) {
	var (
		voters = make([]int32, 0)
		sigs   = make([][]byte, 0)
	)
	// 尝试捕获异常
	defer func() {
		if r := recover(); r != nil {
			ss := s
			log.Println("当前视图信息: ", ss)
			log.Println("捕获到异常: ", r)
			panic("捕获到异常")
		}
	}()
	s.view.mu.Lock()
	defer s.view.mu.Unlock()
	switch msgType {
	case pb2.MsgType_NEW_VIEW:
		for _, vote := range s.view.Vote.NewView {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.NewViewVoter
		return voters, sigs, s.view.once[pb2.MsgType_NEW_VIEW]

	case pb2.MsgType_PREPARE_VOTE:
		for _, vote := range s.view.Vote.Prepare {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.PrepareVoter
		return voters, sigs, s.view.once[pb2.MsgType_PREPARE_VOTE]

	case pb2.MsgType_PRE_COMMIT_VOTE:
		for _, vote := range s.view.Vote.PreCommit {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.PreCommitVoter
		return voters, sigs, s.view.once[pb2.MsgType_PRE_COMMIT_VOTE]

	case pb2.MsgType_COMMIT_VOTE:
		for _, vote := range s.view.Vote.Commit {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.CommitVoter
		return voters, sigs, s.view.once[pb2.MsgType_COMMIT_VOTE]
	}
	return nil, nil, nil
}

func (s *SYNC) GetOnce(megType pb2.MsgType) *dependency.OnceWithDone {
	return s.view.once[megType]
}

func (s *SYNC) HighQC() *pb2.QC {
	// 尝试捕获异常
	defer func() {
		if r := recover(); r != nil {
			ss := s
			log.Println("当前视图信息: ", ss)
			log.Println("捕获到异常: ", r)
			panic("捕获到异常")
		}
	}()
	var highqc = &pb2.QC{}
	for i := 0; i < len(s.view.Vote.NewView); i++ {
		if s.view.Vote.NewView[i].Qc.ViewNumber >= highqc.ViewNumber {
			highqc = s.view.Vote.NewView[i].Qc
		}
	}
	return highqc
}

// 根据存储的投票合成一个QC，已在server中实现
func (s *SYNC) QC(msgType pb2.MsgType) *pb2.QC {
	return &pb2.QC{}
}

func (s *SYNC) MU() *sync.Mutex {
	return &s.mu
}

func (s *SYNC) ViewNumberPP() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentView++
}

func (s *SYNC) ViewNumberSet(v int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentView = v
}

func (s *SYNC) Debug() {
	var (
		sync = modules.MODULES.Synchronizer
		// cryp  = modules.MODULES.SignerAndVerifier
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("同步模块debug info: ", sync, cryp, chain)
}
