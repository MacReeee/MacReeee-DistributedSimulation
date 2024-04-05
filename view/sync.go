package view

import (
	"distributed/modules"
	pb2 "distributed/pb"
	"log"
	"sync"
	"time"
)

var (
	BASE_Timeout = 2 * time.Second //基础超时时间
	MAX_Timeout  = 5 * time.Second //最大超时时间
)

type State int

const (
	Initializing State = iota
	Running

	SuccessSwitching
	TimeoutSwitching
)

type Event int

const (
	StartEvent Event = iota
	TimeoutEvent
	SuccessEvent
)

type SYNC struct {
	State State

	mu   sync.Mutex
	cond *sync.Cond

	CurrentView int64

	view             View          //当前视图
	max              time.Duration // 视图超时的上限
	timeoutMul       time.Duration // 在失败的视图上，将当前平均值乘以此数（应大于1），类似指数退避
	timeoutSwitching bool          // 是否正在切换视图
	SwitchSuccess    chan bool
	ViewSuccess      chan bool

	timer     *time.Timer //每个视图的计时器
	eventChan chan Event

	TimeoutChan chan bool
}

func NewSync() *SYNC {
	s := &SYNC{
		State:       Initializing,
		mu:          sync.Mutex{},
		CurrentView: 0,
		max:         MAX_Timeout,
		timeoutMul:  1,

		TimeoutChan:      make(chan bool, 1),
		eventChan:        make(chan Event),
		timeoutSwitching: false,
		SwitchSuccess:    make(chan bool),
		ViewSuccess:      make(chan bool),
	}
	s.view = *NewView(s)
	s.cond = sync.NewCond(&s.mu)
	go EventLoop(s)
	s.mu.Lock()
	modules.MODULES.Synchronizer = s
	s.mu.Unlock()
	return s
}

func (s *SYNC) Success() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.timeoutSwitching {
		s.State = TimeoutSwitching
		s.SwitchSuccess <- true
	} else {
		s.State = SuccessSwitching
		s.ViewSuccess <- true
	}
	for s.State != Running {
		s.cond.Wait()
	}
}

func (s *SYNC) startView() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State != Initializing {
		// 如果当前状态不是 Initializing，则不应启动新视图
		return
	}
	// 初始化视图运行所需资源
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
	duration := s.view.Duration(s)
	s.timer = time.NewTimer(duration)
	s.State = Running
	s.cond.Broadcast()

	go func() {
		select {
		case <-s.timer.C:
			s.State = TimeoutSwitching
			s.eventChan <- TimeoutEvent
		case <-s.ViewSuccess:
			s.State = SuccessSwitching
			s.eventChan <- SuccessEvent
		}
	}()
}

func (s *SYNC) handleTimeout() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State != TimeoutSwitching {
		return
	}
	s.timeoutSwitching = true

	log.Println("视图 ", s.CurrentView, " 超时")

	// 重置超时倍数
	if s.timeoutMul < MAX_Timeout/BASE_Timeout {
		s.timeoutMul *= 2
	} else {
		// 达到最大倍数时，保持不变或设置为最大超时倍数的值
		s.timeoutMul = MAX_Timeout / BASE_Timeout
	}

	s.view.once[pb2.MsgType_NEW_VIEW].Reset()

	// 通知超时
	s.TimeoutChan <- true

	go func() {
		select {
		case <-s.timer.C:
			s.State = TimeoutSwitching
			s.eventChan <- TimeoutEvent
		case <-s.SwitchSuccess:
			s.timeoutSwitching = false
			s.State = SuccessSwitching
			s.eventChan <- SuccessEvent
		}
	}()
}

func (s *SYNC) handleSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State != SuccessSwitching {
		return
	}
	s.timeoutMul = 1
	s.view = *NewView(s) // 重新初始化视图
	s.State = Initializing
	// 在状态更新后，主动触发 StartEvent，开始新视图的监听
	go func() { s.eventChan <- StartEvent }()
}

func (s *SYNC) Start() {
	s.eventChan <- StartEvent
}

// handleEvent 处理从视图传来的事件，并且根据事件和当前状态来变更状态。
func (s *SYNC) handleEvent(event Event) {

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

// mainEventLoop 是事件循环，负责接收事件并将其传递给状态处理函数。
func EventLoop(s *SYNC) {
	for event := range s.eventChan {
		go s.handleEvent(event)
	}
}

func (s *SYNC) TimerReset() bool {
	duration := s.view.Duration(s)
	//if !s.timer.Stop() {
	//	<-s.timer.C
	//	log.Println("重置计时器")
	//}
	temp := s.timer.Reset(duration)
	return temp
}
