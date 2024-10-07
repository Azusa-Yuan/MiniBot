// Package ymgal 月幕galgame
package ymgal

import (
	database "MiniBot/utils/db"
	"math/rand/v2"
	"strings"

	zero "ZeroBot"
	"ZeroBot/message"
)

func init() {
	engine := zero.NewTemplate(&zero.Metadata{
		Name: "月慕galgame相关",
		Help: "- galCG 随机发一张galCG\n- gal表情包 随机发一张gal表情包\n- galCG[xxx]\n- gal表情包[xxx]\n- 更新gal",
	})
	db := database.GetDefalutDB()
	db.AutoMigrate(&ymgal{})
	gdb = (*ymgaldb)(db)

	engine.OnKeywordGroup([]string{"色图", "涩图", "瑟图"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			y := gdb.randomYmgal(cgType)
			sendYmgal(y, ctx)
		})

	engine.OnRegex(`^gal(CG|表情包)\s*$`).SetBlock(true).
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
	// 这里应该不会有sql注入问题，直接改成除空格符
	engine.OnRegex(`^gal(CG|表情包)\s*(.{1,25})$`).SetBlock(true).
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
	engine.OnFullMatch("更新gal", zero.SuperUserPermission).
		SetBlock(true).SetNoTimeOut(true).Handle(
		func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			err := updatePic()
			if err != nil {
				ctx.SendError(err)
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

	urlList := strings.Split(y.PictureList, ",")
	url := urlList[rand.IntN(len(urlList))]
	ctx.SendChain(message.Text(y.Title), message.Image(url))
}
