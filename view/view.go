package view

import (
	d "distributed/dependency"
	pb2 "distributed/pb"
	"sync"
	"time"
)

type vote struct {
	NewView   []*pb2.NewViewMsg
	Prepare   []*pb2.VoteRequest
	PreCommit []*pb2.VoteRequest
	Commit    []*pb2.VoteRequest

	NewViewVoter   []int32
	PrepareVoter   []int32
	PreCommitVoter []int32
	CommitVoter    []int32
}

type View struct {
	viewnumber int64
	mu         sync.Mutex

	Vote vote                            // 存储投票
	once map[pb2.MsgType]*d.OnceWithDone // 用于保证每个阶段只处理一次投票

	only *sync.Once //用于保证试图成功函数只执行一次

	timer *time.Timer //每个视图的计时器
}

func NewView(s *SYNC) *View {
	v := &View{
		viewnumber: s.CurrentView,
		Vote: vote{
			NewView:   []*pb2.NewViewMsg{},
			Prepare:   []*pb2.VoteRequest{},
			PreCommit: []*pb2.VoteRequest{},
			Commit:    []*pb2.VoteRequest{},

			NewViewVoter:   []int32{},
			PrepareVoter:   []int32{},
			PreCommitVoter: []int32{},
			CommitVoter:    []int32{},
		},
		once: make(map[pb2.MsgType]*d.OnceWithDone),
		only: &sync.Once{},
	}
	v.once[pb2.MsgType_NEW_VIEW] = &d.OnceWithDone{}
	v.once[pb2.MsgType_PREPARE_VOTE] = &d.OnceWithDone{}
	v.once[pb2.MsgType_PRE_COMMIT_VOTE] = &d.OnceWithDone{}
	v.once[pb2.MsgType_COMMIT_VOTE] = &d.OnceWithDone{}

	return v
}

func (v *View) Duration(s *SYNC) time.Duration {
	if s.CurrentView == 0 {
		return 1 * time.Second
	}
	mul := s.timeoutMul
	if BASE_Timeout*mul > MAX_Timeout {
		return MAX_Timeout
	} else {
		return BASE_Timeout * mul
	}
}
