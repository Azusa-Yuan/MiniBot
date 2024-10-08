package monitor

import (
	zero "ZeroBot"
	"fmt"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

var MB = uint64(1024 * 1024)
var memLimit = uint64(150)
var goroutineLimit = 30

func init() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		mem := runtime.MemStats{}
		runtime.ReadMemStats(&mem)
		curMem := mem.Alloc / MB
		if curMem > memLimit {
			select {
			case <-ticker.C:
				zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
					ctx.SendPrivateMessage(zero.BotConfig.GetSuperUser(id)[0], fmt.Sprintf("当前MiniBot内存已达到%d, 请注意", curMem))
					return false
				})
			default:
				log.Warn().Msgf("当前MiniBot内存已达到%d, 请注意", curMem)
			}
		}

		numGoroutine := runtime.NumGoroutine()

		if numGoroutine > goroutineLimit {
			select {
			case <-ticker.C:
				zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
					ctx.SendPrivateMessage(zero.BotConfig.GetSuperUser(id)[0], fmt.Sprintf("当前MiniBot的Goroutine数目已达到%d, 请注意", curMem))
					return false
				})
			default:
				log.Warn().Msgf("当前MiniBot的Goroutine数目已达到%d, 请注意", curMem)
			}
		}

		// 每隔一段时间更新一次内存指标
		time.Sleep(3 * time.Second)
	}()
}
