package dnf

import (
	"MiniBot/plugin/dnf/service"
	zero "ZeroBot"
	"ZeroBot/message"
	"slices"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

func init() {
	engine := zero.NewTemplate(&zero.MetaData{
		Name: "dnf",
		Help: "比例 跨二",
	})

	// 先启动，减少第一次使用的启动时间
	go func() {
		service.Screenshot("跨二", "youxibi")
	}()

	engine.OnPrefixGroup([]string{"比例", "金币", "游戏币"}).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			arg := ctx.State["args"].(string)
			data, url, err := service.Screenshot(arg, "youxibi")
			if err != nil {
				ctx.SendError(err)
			}

			if data == nil {
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.ImageBytes(data), message.Text(url))
		})

	engine.OnPrefixGroup([]string{"矛盾"}).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			arg := ctx.State["args"].(string)
			data, url, err := service.Screenshot(arg, "maodun")
			if err != nil {
				ctx.SendError(err)
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
			users, err := service.GetColgUser()
			if err != nil {
				ctx.SendError(err)
			}
			if ctx.Event.GroupID != 0 {
				groupStr := strconv.FormatInt(ctx.Event.GroupID, 10)
				if slices.Index(users.Group, groupStr) != -1 {
					ctx.SendChain(message.At(ctx.Event.UserID), message.Text("已经订阅过啦"))
					return
				}
				users.Group = append(users.Group, strconv.FormatInt(ctx.Event.GroupID, 10))
			} else {
				qqStr := strconv.FormatInt(ctx.Event.UserID, 10)
				if slices.Index(users.QQ, qqStr) != -1 {
					ctx.SendChain(message.At(ctx.Event.UserID), message.Text("已经订阅过啦"))
					return
				}
				users.QQ = append(users.QQ, qqStr)
			}
			err = users.SaveBinds()
			if err != nil {
				ctx.SendError(err)
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("订阅成功"))
		},
	)

	go func() {
		for range time.NewTicker(180 * time.Second).C {
			users, err := service.GetColgUser()
			if err != nil {
				log.Fatal().Str("name", "dnf").Err(err).Msg("")
			}
			bot, err := zero.GetBot(741433361)
			if err != nil {
				continue
			}
			news, err := users.GetChange()
			if err != nil {
				continue
			}
			for _, new := range news {
				for _, gidStr := range users.Group {
					gid, _ := strconv.ParseInt(gidStr, 10, 64)
					bot.SendGroupMessage(gid, message.Text(new))
				}
				for _, uidStr := range users.QQ {
					uid, _ := strconv.ParseInt(uidStr, 10, 64)
					bot.SendPrivateMessage(uid, message.Text(new))
				}
			}
		}
	}()
}
