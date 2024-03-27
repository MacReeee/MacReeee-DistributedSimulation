package main

import (
	"sync"
	"time"
)

var (
	testmap = make(map[int]int)
	lock    sync.Mutex
)

func testFunc(nums int) {
	res := 1
	for i := 1; i <= nums; i++ {
		res *= i
	}
	lock.Lock()
	defer lock.Unlock()
	testmap[nums] = res
}

func main() {
	for i := 0; i < 10; i++ {
		go testFunc(i)
	}
	time.Sleep(10 * time.Second)
	for key, value := range testmap {
		println(key, value)
	}
}
