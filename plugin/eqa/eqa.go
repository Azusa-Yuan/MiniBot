package eqa

import (
	zero "ZeroBot"

	"github.com/rs/zerolog/log"
)

func init() {
	metaData := zero.MetaData{
		Name: "eqa",
		Help: `### 例子： 设置在默认的情况下
##### 设置一个问题：
- 大家说111回答222
- 我说333回答444
- 大家说@某人回答图1图2 文字
- 大家说图片回答图片
- 有人说R测试回答test`,
	}
	engine := zero.NewTemplate(&metaData)

	engine.OnMessage().SetPriority(11).Handle(
		func(ctx *zero.Ctx) {
			// 测试用
			log.Info().Msg(ctx.Event.RawMessage)

		},
	)
}
