package view

import (
	"context"
	"distributed/hotstuff/pb"
	"sync"
	"time"
)

type view struct {
	Vote vote // 存储投票
	once map[pb.MsgType]*sync.Once

	ctx_success context.Context    //成功的ctx
	success     context.CancelFunc //成功函数

	timer *time.Timer //每个视图的计时器
}

func (v *view) Duration(s *SYNC) time.Duration {
	mul := s.timeoutMul
	if BASE_Timeout*mul > MAX_Timeout {
		return MAX_Timeout
	} else {
		return BASE_Timeout * mul
	}
}

func NewView() *view {
	ctx := context.Background()
	view := &view{
		Vote: vote{
			NewView:   []*pb.NewViewMsg{},
			Prepare:   []*pb.VoteRequest{},
			PreCommit: []*pb.VoteRequest{},
			Commit:    []*pb.VoteRequest{},

			NewViewVoter:   []int32{},
			PrepareVoter:   []int32{},
			PreCommitVoter: []int32{},
			CommitVoter:    []int32{},
		},
		once: make(map[pb.MsgType]*sync.Once),
	}
	view.once[pb.MsgType_NEW_VIEW] = &sync.Once{}
	view.once[pb.MsgType_PREPARE_VOTE] = &sync.Once{}
	view.once[pb.MsgType_PRE_COMMIT_VOTE] = &sync.Once{}
	view.once[pb.MsgType_COMMIT_VOTE] = &sync.Once{}

	view.ctx_success, view.success = context.WithCancel(ctx)
	return view
}

func (v *view) SuccessFunc() context.CancelFunc {
	return v.success
}
