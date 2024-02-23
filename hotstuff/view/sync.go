package view

import (
	"context"
	hotstuff "distributed/hotstuff/consensus"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"log"
	"sync"
	"time"
)

var (
	BASE_Timeout = 500 * time.Millisecond //基础超时时间
	MAX_Timeout  = 10 * time.Second       //最大超时时间
)

// 仅计时相关以及视图切换，不参与共识过程
type Synchronize struct {
	mut         sync.Mutex
	CurrentView int64
	HighQC      *pb.QC //开启一个视图需要一个HighQC
	HighTC      *pb.QC
	duration    ViewDuration
	timer       *time.Timer //每个视图的计时器，超时后打印日志
	
	timeouts    map[int64]map[int32]*pb.TimeoutMsg
}

func New() *Synchronize {
	viewDuration := NewViewDuration(float64(MAX_Timeout), 1)
	Synchronizer := &Synchronize{
		CurrentView: 1,
		HighQC:      nil,
		HighTC:      nil,
		duration:    viewDuration,
		timer:       nil, //超时后打印日志
		timeouts:    make(map[int64]map[int32]*pb.TimeoutMsg),
	}
	modules.MODULES.Synchronizer = Synchronizer
	return Synchronizer
}

func (s *Synchronize) StartTimeOutTimer() {
	s.timer = time.AfterFunc(s.duration.Duration(), func() {
		log.Println("侦测到试图超时")
	})
}

// 启动一个视图，不是整个视图链，ctx是viewDuration中的ctx
func (s *Synchronize) Start(ctx context.Context) {
	s.StartTimeOutTimer()
	//如果视图正常退出，则执行这个协程
	go func() {
		<-ctx.Done()
		s.timer.Stop()
	}()
}

func (s *Synchronize) GetLeader(viewnumber ...int64) int32 {
	if len(viewnumber) == 0 {
		return int32(s.CurrentView)%hotstuff.NumReplicas + 1
	}
	return int32(viewnumber[0])%hotstuff.NumReplicas + 1
}

func (s *Synchronize) TimerReset() bool {
	return s.timer.Reset(s.duration.Duration())
}

func (s *Synchronize) GetContext() context.Context {
	return s.duration.GetContext()
}

type ViewChangeEvent struct {
	View    int64
	Timeout bool
}

type TimeoutEvent struct {
	View int64
}
