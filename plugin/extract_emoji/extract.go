package extractemoji

import (
	zero "ZeroBot"
	"ZeroBot/message"
)

var (
	help = `该服务是针对手机、平板用户的
回复或者引用表情发送 [原图] 或者 [表情提取] 就可以得到该图片的原图
`
	name = "表情提取"
)

func init() {
	engine := zero.NewTemplate(&zero.Metadata{
		Name: name,
		Help: help,
	})
	engine.OnPrefixGroup([]string{"原图", "表情提取"}).Handle(
		func(ctx *zero.Ctx) {
			if ctx.Event.Message[0].Type != "reply" {
				return
			}
			replyMsg := ctx.GetMessage(ctx.Event.Message[0].Data["id"])
			for _, segment := range replyMsg.Elements {
				if segment.Type == "image" {
					url := segment.Data["url"]
					ctx.SendChain(message.Image(url))
					ctx.Stop()
				}
			}
		},
	)
}
