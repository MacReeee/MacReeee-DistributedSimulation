package view

import (
	"context"
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
}

// NewViewDuration 返回一个ViewDuration，它基于先前视图的持续时间来近似视图持续时间。
// 只在初始化副本时调用一次。
// 当发生超时时，下一个视图持续时间将乘以乘数。
// 每个视图就是一个ViewDuration，试图之间不同的的事务定义在这里
func NewViewDuration(maxTimeout, multiplier float64) ViewDuration {
	return &viewDuration{
		max:        maxTimeout,
		timeoutMul: multiplier,
		ctx:        context.Background(),
	}
}

// viewDuration 使用先前视图的统计数据来猜测视图持续时间的合适值。
// 它只考虑有限数量的测量。
type viewDuration struct {
	startTime  time.Time // 当前测量的开始时间
	max        float64   // 视图超时的上限
	timeoutMul float64   // 在失败的视图上，将当前平均值乘以此数（应大于1）
	ctx        context.Context
}

// Duration 返回视图持续时间
func (v *viewDuration) Duration() time.Duration {
	return time.Duration(500*v.timeoutMul) * time.Millisecond
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
	ctx, _ := context.WithTimeout(v.ctx, v.Duration())
	s.Start(ctx)
}

// ViewTimeout 在视图超时时应调用。它将当前平均值乘以'mul'。
func (v *viewDuration) ViewTimeout(s *Synchronize) {
	v.timeoutMul *= 2
	s.CurrentView++
	s.duration = NewViewDuration(v.max, v.timeoutMul)
	ctx, _ := context.WithTimeout(v.ctx, v.Duration())
	s.Start(ctx)
}

func (v *viewDuration) GetContext() context.Context {
	return v.ctx
}
