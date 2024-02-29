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
	MAX_Timeout  = 2 * time.Second        //最大超时时间
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
	// HighQC      *pb.QC //开启一个视图需要一个HighQC
	HighTC   *pb.QC
	duration ViewDuration
	timer    *time.Timer //每个视图的计时器，超时后打印日志 //现行逻辑用不上

	timeouts    map[int64]map[int32]*pb.TimeoutMsg
	TimeoutChan chan bool
}

func New() *Synchronize {
	viewDuration := NewViewDuration(float64(MAX_Timeout), 1)
	// viewDuration.
	Synchronizer := &Synchronize{
		CurrentView: 1,
		// HighQC:      nil,
		// HighTC:      nil,
		duration:    viewDuration,
		timer:       time.NewTimer(1000 * time.Second),
		timeouts:    make(map[int64]map[int32]*pb.TimeoutMsg),
		TimeoutChan: make(chan bool, 1),
	}
	modules.MODULES.Synchronizer = Synchronizer
	return Synchronizer
}

func (s *Synchronize) StartTimeOutTimer(ctx_timeout context.Context, timeout context.CancelFunc) {
	s.timer = time.NewTimer(s.duration.Duration())
	select {
	//////////////////////////////////////////////////////////////////////////超时逻辑
	case <-s.timer.C:
		if ctx_timeout.Err() == nil {
			timeout() //停止正常退出逻辑
			s.duration.ViewTimeout(s)
			s.TimeoutChan <- true //向外部发送超时事件
			log.Println("侦测到试图超时")
		}
	//////////////////////////////////////////////////////////////////////////超时逻辑

	case <-ctx_timeout.Done(): //由视图正常退出触发，等于超时逻辑被取消
		return
	}
}

// 启动一个视图，不是整个视图链，ctx是viewDuration中的ctx
func (s *Synchronize) Start(ctx_success context.Context) {
	ctx_timeout, timeout := context.WithCancel(context.Background())
	go s.StartTimeOutTimer(ctx_timeout, timeout)
	//如果视图正常退出，则执行这个
	select {
	//////////////////////////////////////////////////////////////////////////成功退出视图逻辑
	case <-ctx_success.Done():
		timeout()
		log.Println("视图正常退出")
		s.duration.ViewSucceeded(s)
	//////////////////////////////////////////////////////////////////////////成功退出视图逻辑

	case <-ctx_timeout.Done(): //超时逻辑
		return
	}
}

func (s *Synchronize) GetLeader(viewnumber ...int64) int32 { //如果传入了视图号，则按照传入的视图号计算，否则按照当前视图号计算
	if len(viewnumber) == 0 {
		return int32(s.CurrentView) % hotstuff.NumReplicas
	}
	return int32(viewnumber[0]) % hotstuff.NumReplicas
}

func (s *Synchronize) TimerReset() bool {
	return s.timer.Reset(s.duration.Duration())
}

func (s *Synchronize) GetContext() (context.Context, context.CancelFunc) {
	return s.duration.GetContext(), s.duration.CancelFunc
}

func (s *Synchronize) ViewNumber() int64 {
	return s.CurrentView
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

func (s *Synchronize) GetVoter(msgType pb.MsgType) ([]int32, [][]byte, *sync.Once) {
	voter, sigs := s.duration.GetVoter(msgType)
	return voter, sigs, s.duration.GetOnce(msgType)
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
