package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test_Rand(t *testing.T) {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	test := GenerateExpRand(0.1, r)
	fmt.Println(test)
}

func GenerateExpRand(lambda float64, r *rand.Rand) float64 {
	return r.ExpFloat64() / lambda
}

func Test_time(t *testing.T) {
	originalDuration := 2 * time.Second
	newDuration := 5 * time.Second // 新的持续时间

	timer := time.NewTimer(originalDuration)
	fmt.Println("Original timer started")

	go func() {
		<-timer.C
		fmt.Println("Timer expired")
	}()

	// 假设你决定在原始持续时间结束之前改变持续时间
	time.Sleep(500 * time.Millisecond) // 等待一段时间，然后决定重置计时器
	if !timer.Stop() {
		<-timer.C // 尝试读取通道，确保通道是空的
	}
	timer.Reset(newDuration)
	fmt.Println("Timer reset to new duration")

	// 为了演示，阻塞main函数直到协程完成
	time.Sleep(6 * time.Second)
}
