package manager

import (
	"MiniBot/utils/cache"
	zero "ZeroBot"
	"ZeroBot/message"
)

func init() {
	zero.OnFullMatchGroup([]string{"刷新头像", "更新头像"}).Handle(
		func(ctx *zero.Ctx) {
			_, err := cache.GetAvatarWithoutCache(ctx.Event.UserID)
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("头像刷新啦"))
		},
	)
}
