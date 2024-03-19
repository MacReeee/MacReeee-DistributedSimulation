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

var (
	BASE_Timeout = 5000 * time.Second  //基础超时时间
	MAX_Timeout  = 30000 * time.Second //最大超时时间
)

type State int

const (
	Initializing State = iota
	Running
)

type Event int

const (
	StartEvent Event = iota
	TimeoutEvent
	SuccessEvent
)

type SYNC struct {
	State State

	mu sync.Mutex

	CurrentView int64

	view       view          //当前视图
	max        time.Duration // 视图超时的上限
	timeoutMul time.Duration // 在失败的视图上，将当前平均值乘以此数（应大于1），类似指数退避

	timer     *time.Timer //每个视图的计时器
	eventChan chan Event

	TimeoutChan chan bool
}

// 如果传入视图号，则返回该视图号对应的 Leader 编号，否则返回当前视图对应的 Leader 编号。
func (s *SYNC) GetLeader(viewnumber ...int64) int32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	//todo 测试代码
	//if d.DebugMode {
	//	if len(viewnumber) == 0 {
	//		if s.CurrentView%2 == 0 {
	//			return 2
	//		} else {
	//			return 1
	//		}
	//	}
	//	if viewnumber[0]%2 == 0 {
	//		return 2
	//	} else {
	//		return 1
	//	}
	//}

	if len(viewnumber) == 0 {
		leader := int32(s.CurrentView) % d.NumReplicas
		if leader == 0 {
			return d.NumReplicas
		} else {
			return leader
		}
	}
	leader := int32(viewnumber[0]) % d.NumReplicas
	if leader == 0 {
		return d.NumReplicas
	} else {
		return leader
	}
}

func (s *SYNC) Start() {
	s.eventChan <- StartEvent
}

func (s *SYNC) TimerReset() bool {
	return s.timer.Reset(s.view.Duration(s))
}

func (s *SYNC) GetContext() (context.Context, context.CancelFunc) {
	return s.view.ctx_success, s.view.success
}

func (s *SYNC) ViewNumber() *int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.CurrentView
}

func (s *SYNC) Timeout() <-chan bool {
	return s.TimeoutChan
}

func (s *SYNC) StoreVote(msgType pb.MsgType, NormalMsg *pb.VoteRequest, NewViewMsg ...*pb.NewViewMsg) {
	s.view.mu.Lock()
	defer s.view.mu.Unlock()
	if NormalMsg != nil {
		switch msgType {
		case pb.MsgType_PREPARE_VOTE:
			s.view.Vote.Prepare = append(s.view.Vote.Prepare, NormalMsg)
			s.view.Vote.PrepareVoter = append(s.view.Vote.PrepareVoter, NormalMsg.Voter)
		case pb.MsgType_PRE_COMMIT_VOTE:
			s.view.Vote.PreCommit = append(s.view.Vote.PreCommit, NormalMsg)
			s.view.Vote.PreCommitVoter = append(s.view.Vote.PreCommitVoter, NormalMsg.Voter)
		case pb.MsgType_COMMIT_VOTE:
			s.view.Vote.Commit = append(s.view.Vote.Commit, NormalMsg)
			s.view.Vote.CommitVoter = append(s.view.Vote.CommitVoter, NormalMsg.Voter)
		}
	}
	if NewViewMsg != nil {
		s.view.Vote.NewView = append(s.view.Vote.NewView, NewViewMsg...)
		s.view.Vote.NewViewVoter = append(s.view.Vote.NewViewVoter, NewViewMsg[0].ProposalId)
	}
}

func (s *SYNC) GetVoter(msgType pb.MsgType) ([]int32, [][]byte, *d.OnceWithDone) {
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
	case pb.MsgType_NEW_VIEW:
		for _, vote := range s.view.Vote.NewView {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.NewViewVoter
		return voters, sigs, s.view.once[pb.MsgType_NEW_VIEW]

	case pb.MsgType_PREPARE_VOTE:
		for _, vote := range s.view.Vote.Prepare {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.PrepareVoter
		return voters, sigs, s.view.once[pb.MsgType_PREPARE_VOTE]

	case pb.MsgType_PRE_COMMIT_VOTE:
		for _, vote := range s.view.Vote.PreCommit {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.PreCommitVoter
		return voters, sigs, s.view.once[pb.MsgType_PRE_COMMIT_VOTE]

	case pb.MsgType_COMMIT_VOTE:
		for _, vote := range s.view.Vote.Commit {
			sigs = append(sigs, vote.Signature)
		}
		voters = s.view.Vote.CommitVoter
		return voters, sigs, s.view.once[pb.MsgType_COMMIT_VOTE]
	}
	return nil, nil, nil
}

func (s *SYNC) GetOnce(megType pb.MsgType) *d.OnceWithDone {
	return s.view.once[megType]
}

func (s *SYNC) HighQC() *pb.QC {
	// 尝试捕获异常
	defer func() {
		if r := recover(); r != nil {
			ss := s
			log.Println("当前视图信息: ", ss)
			log.Println("捕获到异常: ", r)
			panic("捕获到异常")
		}
	}()
	var highqc = &pb.QC{}
	for i := 0; i < len(s.view.Vote.NewView); i++ {
		if s.view.Vote.NewView[i].Qc.ViewNumber > highqc.ViewNumber {
			highqc = s.view.Vote.NewView[i].Qc
		}
	}
	return highqc
}

// 根据存储的投票合成一个QC，已在server中实现
func (s *SYNC) QC(msgType pb.MsgType) *pb.QC {
	return &pb.QC{}
}

func (s *SYNC) MU() *sync.Mutex {
	return &s.mu
}

func (s *SYNC) ViewNumberPP() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentView++
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

func NewSync() *SYNC {
	sync := &SYNC{
		State:       Initializing,
		mu:          sync.Mutex{},
		CurrentView: 0,
		view:        *NewView(),
		max:         MAX_Timeout,
		timeoutMul:  1,

		TimeoutChan: make(chan bool, 1),
		eventChan:   make(chan Event),
	}
	go EventLoop(sync)
	sync.mu.Lock()
	modules.MODULES.Synchronizer = sync
	sync.mu.Unlock()
	return sync
}

// handleEvent 处理从视图传来的事件，并且根据事件和当前状态来变更状态。
func (s *SYNC) handleEvent(event Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch event {
	case StartEvent:
		s.startView()
	case TimeoutEvent:
		s.handleTimeout()
	case SuccessEvent:
		s.handleSuccess()
		// 可能还有其他状态和事件的处理
	}
}

func (s *SYNC) startView() {
	if s.State != Initializing {
		// 如果当前状态不是 Initializing，则不应启动新视图
		return
	}
	s.State = Running
	// 初始化视图运行所需资源
	//log.Println("开启新视图，当前视图: ", s.CurrentView)
	// 设置超时定时器
	if s.timer != nil {
		if !s.timer.Stop() { // 确保停止旧的计时器
			select {
			case <-s.timer.C:
				// 定时器已过期，从通道中成功读取
			default:
				// 定时器尚未过期，或者已经被读取，不做任何操作
			}
		}
	}
	s.timer = time.NewTimer(s.view.Duration(s))

	go func() {
		select {
		case <-s.view.ctx_success.Done():
			s.eventChan <- SuccessEvent
		case <-s.timer.C:
			s.eventChan <- TimeoutEvent
		}
	}()
}

func (s *SYNC) handleTimeout() {
	if s.State != Running {
		// 只有在 Running 状态时，超时才有效
		return
	}
	s.TimeoutChan <- true
	log.Println("视图 ", s.CurrentView, " 超时")
	if s.timeoutMul < MAX_Timeout/BASE_Timeout {
		s.timeoutMul *= 2
	} else {
		// 达到最大倍数时，保持不变或设置为最大超时倍数的值
		s.timeoutMul = MAX_Timeout / BASE_Timeout
	}
	s.prepareForNextView()
}

func (s *SYNC) handleSuccess() {
	if s.State != Running {
		// 只有在 Running 状态时，成功才有效
		return
	}
	// log.Println("视图 ", s.CurrentView, " 成功退出")
	s.timeoutMul = 1
	s.prepareForNextView()
}

func (s *SYNC) prepareForNextView() {
	s.view = *NewView() // 重新初始化视图
	s.State = Initializing
	// 在状态更新后，主动触发 StartEvent，开始新视图的监听
	go func() { s.eventChan <- StartEvent }()
}

// mainEventLoop 是事件循环，负责接收事件并将其传递给状态处理函数。
func EventLoop(s *SYNC) {
	for event := range s.eventChan {
		s.handleEvent(event)
	}
}
