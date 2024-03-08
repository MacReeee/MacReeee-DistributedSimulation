package view

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func ViewSuccess(sync *Synchronize) {
	_, success := sync.GetContext()
	success()
}

func TestSynchronize_Start(t *testing.T) {
	sync := New() // 假设 New 函数返回 *Synchronize 实现了 Synchronizer 接口
	_, success := sync.GetContext()

	// 启动视图管理器
	go sync.Start()

	//启动超时事件接收器
	go func() {
		timer := time.After(15 * time.Second)
		i := 1
		for {
			select {
			case <-timer:
				return
			case <-sync.Timeout():
				fmt.Printf("侦测到超时事件: %d, 当前视图号: %d\n", i, sync.CurrentView)
				i++
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
	currentView := *sync.ViewNumber()
	ViewSuccess(sync)
	// expectedLeader := int32(currentView)%4 + 1
	// leader := sync.GetLeader()

	// if leader != expectedLeader {
	// 	t.Errorf("GetLeader() = %v, want %v", leader, expectedLeader)
	// }

	// 模拟工作一段时间后结束视图
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)
	_, success = sync.GetContext()
	success()
	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Second)

	// 验证视图编号是否增加
	newViewNumber := sync.ViewNumber()
	log.Println("最终视图号: ", sync.CurrentView)
	if *newViewNumber <= currentView {
		t.Errorf("ViewNumber did not increase after view succeeded")
	}
}

// 更多的测试...

// func Test_ViewSuccess(t *testing.T) {
// 	sync := New() // 假设 New 函数返回 *Synchronize 实现了 Synchronizer 接口
// 	ctx_success, success := sync.GetContext()
// }

func Test_NewSync(t *testing.T) {
	sync := NewSync()
	go func() {
		for {
			<-sync.Timeout()
		}
	}()
	_, success := sync.GetContext()
	log.Println("当前视图号: ", sync.ViewNumber())

	sync.Start()

	time.Sleep(5 * time.Second)
	success()

	time.Sleep(1000 * time.Second)
}

func Test_GetLeader(t *testing.T) {
	sync := NewSync()
	for i := 0; i <= 20; i++ {
		log.Println("输入视图 ", i, " 的主节点为: ", sync.GetLeader(int64(i)))
	}
}
