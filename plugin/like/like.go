package like

import (
	zero "ZeroBot"
	"ZeroBot/message"
)

var (
	pluginName = "like"
)

func init() {
	metaData := &zero.MetaData{
		Name: pluginName,
		Help: "点赞",
	}
	engine := zero.NewTemplate(metaData)
	engine.OnFullMatch("赞我", zero.OnlyToMe).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			err := ctx.SendLike(ctx.Event.UserID, 10)
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("好啦好啦,已经点赞10次啦"))
		},
	)
}
