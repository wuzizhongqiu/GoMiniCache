package wait

/**
 * 自己封装的 wait，主要实现了一个超时的功能（因为 go 语言提供的包是没有超时功能的）
 */

import (
	"sync"
	"time"
)

// Wait 跟库里的 sync.WaitGroup 基本一致，额外实现了超时功能
type Wait struct {
	wg sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

func (w *Wait) Done() {
	w.wg.Done()
}

func (w *Wait) Wait() {
	w.wg.Wait() // 阻塞，直到计数器为 0
}

// WaitWithTimeout 会阻塞，直到计数器为 0，或者超时
// 如果出现超时，返回 true
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan bool, 1)
	go func() {
		defer close(c)
		w.wg.Wait()
		c <- true
	}()
	select {
	case <-c:
		return false // 正常完成
	case <-time.After(timeout):
		return true // 出现超时
	}
}
