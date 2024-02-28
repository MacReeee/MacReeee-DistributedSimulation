package view

import (
	"fmt"
	"testing"
	"time"
)

func ViewSuccess(sync *Synchronize) {
	_, success := sync.GetContext()
	success()
}

func TestSynchronize_Start(t *testing.T) {
	sync := New() // 假设 New 函数返回 *Synchronize 实现了 Synchronizer 接口
	ctx, success := sync.GetContext()

	// 启动视图管理器
	go sync.Start(ctx)

	//启动超时事件接收器
	go func() {
		timer := time.After(15 * time.Second)
		for {
			select {
			case <-timer:
				return
			case <-sync.Timeout():
				fmt.Printf("侦测到超时事件\n")
			}
		}
	}()
	fmt.Println("test")
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	ViewSuccess(sync)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	ViewSuccess(sync)
	time.Sleep(1 * time.Second)
	ViewSuccess(sync)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)

	// 假设 Leader 的计算是基于当前视图编号的
	currentView := sync.ViewNumber()
	ViewSuccess(sync)
	expectedLeader := int32(currentView)%4 + 1
	leader := sync.GetLeader()

	if leader != expectedLeader {
		t.Errorf("GetLeader() = %v, want %v", leader, expectedLeader)
	}

	// 模拟工作一段时间后结束视图
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	ctx, success = sync.GetContext()
	success()
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)

	// 验证视图编号是否增加
	newViewNumber := sync.ViewNumber()
	if newViewNumber <= currentView {
		t.Errorf("ViewNumber did not increase after view succeeded")
	}
}

// 更多的测试...
