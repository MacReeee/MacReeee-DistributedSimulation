package view

import (
	"context"
	"distributed/hotstuff/pb"
	"fmt"
	"time"
)

// ViewDuration 确定视图的持续时间。
// 视图同步器使用此接口设置其超时时间。
type ViewDuration interface {
	Duration() time.Duration    // Duration 返回下一个视图应持续的时间。
	ViewStarted()               // ViewStarted 在启动新视图时由同步器调用。
	ViewSucceeded(*Synchronize) // ViewSucceeded 在视图成功结束时由同步器调用。
	ViewTimeout(*Synchronize)   // ViewTimeout 在视图超时时由同步器调用。
	GetContext() context.Context

	/* -------导出类型------- */
	CancelFunc()
	vote() *vote
	GetVoter(msgType pb.MsgType) ([]int32, [][]byte)
	HighQC() *pb.QC
	QC(msgType pb.MsgType) *pb.QC // 根据存储的投票合成一个QC
}

type vote struct {
	NewView   []*pb.NewViewMsg
	Prepare   []*pb.VoteRequest
	PreCommit []*pb.VoteRequest
	Commit    []*pb.VoteRequest

	NewViewVoter   []int32
	PrepareVoter   []int32
	PreCommitVoter []int32
	CommitVoter    []int32
}

// viewDuration 使用先前视图的统计数据来猜测视图持续时间的合适值。
// 它只考虑有限数量的测量。
type viewDuration struct {
	startTime  time.Time // 当前测量的开始时间
	max        float64   // 视图超时的上限
	timeoutMul float64   // 在失败的视图上，将当前平均值乘以此数（应大于1）
	ctx        context.Context
	success    context.CancelFunc

	Vote vote
}

// NewViewDuration 返回一个ViewDuration，它基于先前视图的持续时间来近似视图持续时间。
// 只在初始化副本时调用一次。
// 当发生超时时，下一个视图持续时间将乘以乘数。
// 每个视图就是一个ViewDuration，试图之间不同的的事务定义在这里
func NewViewDuration(maxTimeout, multiplier float64) ViewDuration {
	ctx := context.Background()
	view := &viewDuration{
		max:        maxTimeout,
		timeoutMul: multiplier,
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
	}
	view.ctx, view.success = context.WithCancel(ctx)
	fmt.Println("viewDuration")
	return view
}

func (v *viewDuration) CancelFunc() {
	v.success()
}

// Duration 返回视图持续时间
func (v *viewDuration) Duration() time.Duration {
	if BASE_Timeout*time.Duration(v.timeoutMul) > MAX_Timeout {
		return MAX_Timeout
	} else {
		return BASE_Timeout * time.Duration(v.timeoutMul)
	}
}

// ViewStarted 记录视图的开始时间。
func (v *viewDuration) ViewStarted() {
	v.startTime = time.Now()
}

// ViewSucceeded 计算视图的持续时间
// 并更新用于计算平均值和方差的内部值。
func (v *viewDuration) ViewSucceeded(s *Synchronize) {
	v.timeoutMul = 1
	s.CurrentView++
	s.duration = NewViewDuration(v.max, v.timeoutMul)
	s.Start(v.GetContext())
}

// ViewTimeout 在视图超时时应调用。它将当前平均值乘以'mul'。
func (v *viewDuration) ViewTimeout(s *Synchronize) {
	v.timeoutMul *= 2
	s.CurrentView++
	s.duration = NewViewDuration(v.max, v.timeoutMul)
	s.Start(v.GetContext())
}

func (v *viewDuration) GetContext() context.Context {
	return v.ctx
}

func (v *viewDuration) vote() *vote {
	return &v.Vote
}

func (v *viewDuration) GetVoter(msgType pb.MsgType) ([]int32, [][]byte) {
	var sigs [][]byte
	switch msgType {
	case pb.MsgType_NEW_VIEW:
		for _, vote := range v.Vote.NewView {
			sigs = append(sigs, vote.Signature)
		}
		return v.Vote.NewViewVoter, sigs

	case pb.MsgType_PREPARE_VOTE:
		for _, vote := range v.Vote.Prepare {
			sigs = append(sigs, vote.Signature)
		}
		return v.Vote.PrepareVoter, sigs

	case pb.MsgType_PRE_COMMIT_VOTE:
		for _, vote := range v.Vote.PreCommit {
			sigs = append(sigs, vote.Signature)
		}
		return v.Vote.PreCommitVoter, sigs

	case pb.MsgType_COMMIT_VOTE:
		for _, vote := range v.Vote.Commit {
			sigs = append(sigs, vote.Signature)
		}
	}
	return nil, nil
}

func (v *viewDuration) HighQC() *pb.QC {
	var highqc *pb.QC = v.Vote.NewView[0].Qc
	for i := 1; i < len(v.Vote.NewView); i++ {
		if v.Vote.NewView[i].Qc.ViewNumber > highqc.ViewNumber {
			highqc = v.Vote.NewView[i].Qc
		}
	}
	return highqc
}

func (v *viewDuration) QC(msgType pb.MsgType) *pb.QC {
	QC := &pb.QC{}
	// var votes *vote

	// switch msgType {
	// case pb.MsgType_PREPARE_VOTE:
	// 	votes
	// }
	// var votes []*pb.VoteRequest = v.

	return QC
}
