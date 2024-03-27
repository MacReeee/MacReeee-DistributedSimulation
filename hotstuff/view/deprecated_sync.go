// Deprecated: 该同步器实现已弃用
// 由于出现未知原因的内存泄漏，该模块及其对应的视图模块viewduration.go已被弃用
// 使用同包下的sync.go和view.go替代

package view

import (
	"context"
	d "distributed/hotstuff/dependency"
	"distributed/hotstuff/modules"
	"distributed/hotstuff/pb"
	"log"
	"sync"
	"time"
)

// /* ------- debug mode ------- */
// var (
// 	BASE_Timeout = 500 * time.Second  //基础超时时间
// 	MAX_Timeout  = 2000 * time.Second //最大超时时间
// )

// /* ----- debug mode end ----- */

// 仅计时相关以及视图切换，不参与共识过程
type Synchronize struct {
	mu          sync.Mutex
	CurrentView int64
	// HighQC   *pb.QC //开启一个视图需要一个HighQC
	// HighTC   *pb.QC
	duration ViewDuration
	timer    *time.Timer //每个视图的计时器，超时后打印日志 //现行逻辑用不上

	timeouts    map[int64]map[int32]*pb.TimeoutMsg
	TimeoutChan chan bool

	debug_count int
}

func (s *Synchronize) ViewNumberPP() {
	panic("已弃用")
}

func (s *Synchronize) MU() *sync.Mutex {
	panic("已弃用")
}

func (s *Synchronize) GetOnce(megType pb.MsgType) *d.OnceWithDone {
	//不用实现，已经废弃
	panic("已弃用")
}

// Deprecated: 该同步器实现已弃用
func New() *Synchronize {
	viewDuration := NewViewDuration(float64(MAX_Timeout), 1)
	// viewDuration.
	Synchronizer := &Synchronize{
		CurrentView: 0,
		// HighQC:      nil,
		// HighTC:      nil,
		duration:    viewDuration,
		timer:       time.NewTimer(1000 * time.Second),
		timeouts:    make(map[int64]map[int32]*pb.TimeoutMsg),
		TimeoutChan: make(chan bool, 1),
	}
	Synchronizer.mu.Lock()
	modules.MODULES.Synchronizer = Synchronizer
	Synchronizer.mu.Unlock()
	return Synchronizer
}

// 启动一个视图，不是整个视图链，ctx是viewDuration中的ctx
func (s *Synchronize) Start() {
	log.Println(s.CurrentView)
	ctx_success := s.duration.GetContext()
	s.timer = time.NewTimer(s.duration.Duration())
	//如果视图正常退出，则执行这个
	select {
	case <-ctx_success.Done():
		log.Println("视图正常退出")
		s.duration.ViewSucceeded(s)
		return

	//如果视图超时，则执行这个
	case <-s.timer.C:
		s.duration.ViewTimeout(s)
		s.mu.Lock()
		s.debug_count++
		s.mu.Unlock()
		log.Println("侦测到试图超时, 已侦测到", s.debug_count, "次超时事件")
		s.TimeoutChan <- true //向外部发送超时事件
		return
	}
}

func (s *Synchronize) GetLeader(viewnumber ...int64) int32 { //如果传入了视图号，则按照传入的视图号计算，否则按照当前视图号计算
	if len(viewnumber) == 0 {
		s.mu.Lock()
		defer s.mu.Unlock()
		return int32(s.CurrentView) % d.NumReplicas
	}
	return int32(viewnumber[0]) % d.NumReplicas
}

func (s *Synchronize) TimerReset() bool {
	return s.timer.Reset(s.duration.Duration())
}

func (s *Synchronize) GetContext() (context.Context, context.CancelFunc) {
	return s.duration.GetContext(), s.duration.SuccessFunc()
}

func (s *Synchronize) GetOnly() *sync.Once {
	panic("已弃用")
}

func (s *Synchronize) ViewNumber() *int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.CurrentView
}

func (s *Synchronize) Timeout() <-chan bool {
	return s.TimeoutChan
}

func (s *Synchronize) StoreVote(msgType pb.MsgType, NormalMsg *pb.VoteRequest, NewViewMsg ...*pb.NewViewMsg) {
	if NormalMsg != nil {
		switch msgType {
		case pb.MsgType_PREPARE_VOTE:
			s.duration.vote().Prepare = append(s.duration.vote().Prepare, NormalMsg)
		case pb.MsgType_PRE_COMMIT_VOTE:
			s.duration.vote().PreCommit = append(s.duration.vote().PreCommit, NormalMsg)
		case pb.MsgType_COMMIT_VOTE:
			s.duration.vote().Commit = append(s.duration.vote().Commit, NormalMsg)
		}
	}
	if NewViewMsg != nil {
		s.duration.vote().NewView = append(s.duration.vote().NewView, NewViewMsg...)
	}
}

func (s *Synchronize) GetVoter(msgType pb.MsgType) ([]int32, [][]byte, *d.OnceWithDone) {
	voter, sigs := s.duration.GetVoter(msgType)
	once := s.duration.GetOnce(msgType)
	return voter, sigs, once
}

func (s *Synchronize) HighQC() *pb.QC {
	return s.duration.HighQC()
}

func (s *Synchronize) QC(msgType pb.MsgType) *pb.QC {
	return s.duration.QC(msgType)
}

/* -------utils functions------- */

func (s *Synchronize) Vote() *vote {
	return s.duration.vote()
}

func (s *Synchronize) Debug() {
	var (
		sync = modules.MODULES.Synchronizer
		// cryp  = modules.MODULES.SignerAndVerifier
		cryp  = modules.MODULES.Signer
		chain = modules.MODULES.Chain
	)
	log.Println("同步模块debug info: ", sync, cryp, chain)
}

/* -----utils functions end----- */

// type ViewChangeEvent struct {
// 	View    int64
// 	Timeout bool
// }

// type TimeoutEvent struct {
// 	View int64
// }
