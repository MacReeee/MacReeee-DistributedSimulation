package view

import (
	"context"
	"distributed/dependency"
	"distributed/modules"
	pb2 "distributed/pb"
	"log"
	"sync"
	"time"
)

var (
	BASE_Timeout = 20000 * time.Second //基础超时时间
	MAX_Timeout  = 50000 * time.Second //最大超时时间
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
	ctx       context.Context

	TimeoutChan chan bool
}

func NewSync(ctx context.Context) *SYNC {
	s := &SYNC{
		State:       Initializing,
		mu:          sync.Mutex{},
		CurrentView: 0,
		max:         dependency.GetMAX_Timeout(),
		timeoutMul:  1,

		TimeoutChan:      make(chan bool, 1),
		eventChan:        make(chan Event),
		timeoutSwitching: false,
		SwitchSuccess:    make(chan bool),
		ViewSuccess:      make(chan bool),
		ctx:              ctx,
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
	if s.timeoutSwitching { //如果发生了超时，成功切换视图以后走此分支
		s.State = TimeoutSwitching
		s.SwitchSuccess <- true
	} else { //正常情况下的分支
		s.State = SuccessSwitching
		s.ViewSuccess <- true
	}
	for s.State != Running { //等待新视图创建并初始化成功
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
	if s.timeoutMul < dependency.GetMAX_Timeout()/dependency.GetBASE_Timeout() {
		s.timeoutMul *= 2
	} else {
		// 达到最大倍数时，保持不变或设置为最大超时倍数的值
		s.timeoutMul = dependency.GetMAX_Timeout() / dependency.GetBASE_Timeout()
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
	for {
		select {
		case event := <-s.eventChan:
			go s.handleEvent(event)
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *SYNC) TimerReset() bool {
	duration := s.view.Duration(s)
	temp := s.timer.Reset(duration)
	return temp
}
