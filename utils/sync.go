package cherryUtils

import goSync "sync"

// WaitGroup 封装 sync.WaitGroup，提供更简单的 API
type WaitGroup struct {
	wg goSync.WaitGroup
}

// Add 添加一个非阻塞的任务，任务在新的 Go 程执行
func (wg *WaitGroup) Add(fn func()) {
	wg.wg.Add(1)
	go func() {
		defer wg.wg.Done()
		fn()
	}()
}

// Wait 等待所有任务执行完成
func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}
