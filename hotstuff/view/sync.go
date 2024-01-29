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
	currentView int64
	highQC      *pb.QC //开启一个视图需要一个HighQC
	highTC      *pb.QC
	duration    ViewDuration
	timer       *time.Timer //是一个log记录器，在计时器结束后执行日志记录等相关函数，如果有需要超时后执行的函数，可以使用这个
	timeouts    map[int64]map[int32]*pb.TimeoutMsg
}

func New(viewDuration ViewDuration) *Synchronize {
	Synchronizer := &Synchronize{
		currentView: 1,
		highQC:      nil,
		highTC:      nil,
		duration:    viewDuration,
		timer:       nil,
		timeouts:    make(map[int64]map[int32]*pb.TimeoutMsg),
	}
	modules.MODULES.Synchronizer = Synchronizer
	return Synchronizer
}

func (s *Synchronize) startTimeOutTimer() {
	s.timer = time.AfterFunc(s.duration.Duration(), func() {
		log.Println("侦测到试图超时")
	})
}

// ctx是整个视图链的ctx
func (s *Synchronize) Start(ctx context.Context) {
	s.startTimeOutTimer()

	//如果视图正常退出，则执行这个协程
	go func() {
		<-ctx.Done()
		s.timer.Stop()
	}()

	//启动整个共识过程
}

func (s *Synchronize) GetLeader(viewnumber int64) int32 {
	return int32(viewnumber)%hotstuff.NumReplicas + 1
}

type ViewChangeEvent struct {
	View    int64
	Timeout bool
}

type TimeoutEvent struct {
	View int64
}
