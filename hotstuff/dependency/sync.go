package dependency

import "sync"

type OnceWithDone struct {
	once sync.Once
	done bool
	mu   sync.Mutex // 保护done字段
}

func (o *OnceWithDone) Do(f func()) {
	o.once.Do(func() {
		f()
		o.mu.Lock()
		o.done = true
		o.mu.Unlock()
	})
}

func (o *OnceWithDone) IsDone() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.done
}
