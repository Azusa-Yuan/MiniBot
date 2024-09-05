package utils

import "sync"

// GlobalInitMutex 在 init 时被冻结, main 初始化完成后解冻
var GlobalInitMutex = func() (mu sync.Mutex) {
	mu.Lock()
	return
}()
