package manager

import (
	"MiniBot/plugin/manager/plugin"
	zero "ZeroBot"
	"ZeroBot/extension"
	"ZeroBot/message"
	"fmt"
	"strconv"
)

func init() {
	zero.OnPrefixGroup([]string{"用法", "usage", "帮助", "help"}, zero.OnlyToMe).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			model := extension.PrefixModel{}
			_ = ctx.Parse(&model)
			service, ok := plugin.CM.Lookup(model.Args)
			if !ok {
				ctx.SendChain(message.Text("没有找到指定服务!"))
				return
			}
			if service.String() == "" {
				ctx.SendChain(message.Text("该服务无帮助!"))
				return
			}
			// Todo未查看是否有足够权限
			ctx.SendChain(message.Text(service.String()))
		})

	zero.OnFullMatchGroup([]string{"插件列表", "lssv", "服务列表"}, zero.OnlyToMe).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			bid := ctx.Event.SelfID
			uidKey := bidWithuid(bid, uid)
			gidKey := bidWithgid(bid, gid)
			bidKey := strconv.FormatInt(bid, 10)
			if managers.IsBlocked(uidKey) || managers.IsBlocked(gidKey) {
				resp := fmt.Sprintf(zero.BotConfig.NickName[0], "睡着了")
				ctx.SendChain(message.Text(resp))
				return
			}
			resp := "您可以使用的服务如下:"
			plugin.CM.RLock()
			defer plugin.CM.RUnlock()
			for _, c := range plugin.CM.ControlMap {
				if c.IsEnabled("0") && c.IsEnabled(bidKey) && c.IsEnabled(gidKey) && c.IsEnabled(uidKey) {
					resp += "\n" + c.MetaDate.Name
				}
			}
			resp += "\n发送\"帮助 服务名字\"查看服务帮助"
			ctx.SendChain(message.Text(resp))
		})
}
