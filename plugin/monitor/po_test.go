package monitor

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestXxx(t *testing.T) {
	ticker := time.NewTicker(5 * time.Second)
	time.Sleep(10 * time.Second)
	for {
		// 获取当前内存使用情况
		// 这里可以使用 runtime 包获取 Go 程序的内存使用情况
		// 例如：mem := runtime.MemStats{}
		// runtime.ReadMemStats(&mem)
		// memoryUsage.Set(float64(mem.Alloc))

		// 模拟内存使用情况的更新
		mem := runtime.MemStats{}
		runtime.ReadMemStats(&mem)
		select {
		case <-ticker.C:
			fmt.Println(mem.Alloc)
		default:
		}
		fmt.Println("通过了")

		// 每隔一段时间更新一次内存指标
		time.Sleep(1 * time.Second)
	}
}
