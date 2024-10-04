package dnf

import (
	"MiniBot/plugin/dnf/service"
	"MiniBot/service/book"
	zero "ZeroBot"
	"ZeroBot/message"
	"time"

	"github.com/rs/zerolog/log"
)

func init() {
	engine := zero.NewTemplate(&zero.MetaData{
		Name: "dnf",
		Help: `查询DNF金币或者矛盾价格  实例指令 比例跨二  矛盾跨二
查看colg资讯 指令 colg资讯
订阅colg资讯 指令 订阅colg资讯`,
	})

	engine.OnPrefixGroup([]string{"比例", "金币", "游戏币"}, IfSever).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(zero.BotConfig.GetNickName(ctx.Event.SelfID)[0] + "在查了在查了。。。"))
			arg := ctx.State["args"].(string)
			data, url, err := service.Screenshot(arg, "youxibi")
			if err != nil {
				ctx.SendError(err, message.Text("网络不稳定喵，请点链接查看"), message.Text(url))
			}

			if data == nil {
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.ImageBytes(data), message.Text(url))
		})

	engine.OnPrefixGroup([]string{"矛盾"}, IfSever).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(zero.BotConfig.GetNickName(ctx.Event.SelfID)[0] + "在查了在查了。。。"))
			arg := ctx.State["args"].(string)
			data, url, err := service.Screenshot(arg, "maodun")
			if err != nil {
				ctx.SendError(err, message.Text("网络不稳定喵，请点链接查看"), message.Text(url))
			}

			if data == nil {
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.ImageBytes(data), message.Text(url))
		})

	engine.OnFullMatch("colg资讯").SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			contend, err := service.ColgNews()
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(contend))
		})

	engine.OnFullMatch("订阅colg资讯", zero.UserOrGrpAdmin).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			if ctx.Event.GroupID != 0 {
				uid = 0
			}
			err := book.CreatOrUpdateBookInfo(&book.Book{
				BotID:   ctx.Event.SelfID,
				UserID:  uid,
				GroupID: ctx.Event.GroupID,
				Service: "colg",
			})

			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("订阅成功"))
		},
	)

	engine.OnFullMatch("取消订阅colg资讯", zero.UserOrGrpAdmin).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			if ctx.Event.GroupID != 0 {
				uid = 0
			}
			err := book.DeleteBookInfo(&book.Book{
				BotID:   ctx.Event.SelfID,
				UserID:  uid,
				GroupID: ctx.Event.GroupID,
				Service: "colg",
			})

			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("删除订阅成功"))
		},
	)

	go func() {
		for range time.NewTicker(180 * time.Second).C {
			news, err := service.GetColgChange()
			if len(news) == 0 {
				if err != nil {
					log.Error().Str("name", "dnf").Err(err).Msg("")
				}
				continue
			}

			bookInfos, err := book.GetBookInfos("colg")
			if err != nil {
				log.Error().Str("name", "dnf").Err(err).Msg("")
			}

			for _, new := range news {
				for _, bookInfo := range bookInfos {
					bot, err := zero.GetBot(bookInfo.BotID)
					if err != nil {
						continue
					}
					if bookInfo.GroupID != 0 {
						bot.SendGroupMessage(bookInfo.GroupID, message.Text(new))
					} else {
						bot.SendPrivateMessage(bookInfo.UserID, message.Text(new))
					}
				}
			}
		}
	}()
}

func IfSever(ctx *zero.Ctx) bool {
	_, ok := service.ReportRegions[ctx.State["args"].(string)]
	return ok
}
