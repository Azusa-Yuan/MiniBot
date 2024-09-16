// Package ymgal 月幕galgame
package ymgal

import (
	database "MiniBot/utils/db"
	"MiniBot/utils/message_tools"
	"strings"

	zero "ZeroBot"
	"ZeroBot/message"
)

func init() {
	engine := zero.NewTemplate(&zero.MetaData{
		Name: "月慕galgame相关",
		Help: "- 随机galCG\n- 随机gal表情包\n- galCG[xxx]\n- gal表情包[xxx]\n- 更新gal",
	})
	db := database.DbConfig.GetDb("lulumu")
	db.AutoMigrate(&ymgal{})
	gdb = (*ymgaldb)(db)

	engine.OnRegex("^随机gal(CG|表情包)$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			pictureType := ctx.State["regex_matched"].([]string)[1]
			var y ymgal
			if pictureType == "表情包" {
				y = gdb.randomYmgal(emoticonType)
			} else {
				y = gdb.randomYmgal(cgType)
			}
			sendYmgal(y, ctx)
		})
	engine.OnRegex("^gal(CG|表情包)([一-龥ぁ-んァ-ヶA-Za-z0-9]{1,25})$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			pictureType := ctx.State["regex_matched"].([]string)[1]
			key := ctx.State["regex_matched"].([]string)[2]
			var y ymgal
			if pictureType == "CG" {
				y = gdb.getYmgalByKey(cgType, key)
			} else {
				y = gdb.getYmgalByKey(emoticonType, key)
			}
			sendYmgal(y, ctx)
		})
	engine.OnFullMatch("更新gal", zero.SuperUserPermission).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			err := updatePic()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send("ymgal数据库已更新")
		})
}

func sendYmgal(y ymgal, ctx *zero.Ctx) {
	if y.PictureList == "" {
		ctx.SendChain(message.Text(zero.BotConfig.NickName[0] + "暂时没有这样的图呢"))
		return
	}
	m := message.Message{message_tools.FakeSenderForwardNode(ctx, message.Text(y.Title))}
	if y.PictureDescription != "" {
		m = append(m, message_tools.FakeSenderForwardNode(ctx, message.Text(y.PictureDescription)))
	}
	for _, v := range strings.Split(y.PictureList, ",") {
		m = append(m, message_tools.FakeSenderForwardNode(ctx, message.Image(v)))
	}
	if id := ctx.Send(m).ID(); id == 0 {
		ctx.SendChain(message.Text("ERROR: 可能被风控了"))
	}
}
