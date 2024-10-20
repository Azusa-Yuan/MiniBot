// 部分参考 Package picturePackage 月幕galgame
package picturepackage

import (
	database "MiniBot/utils/db"
	"math/rand/v2"
	"regexp"
	"strings"

	zero "ZeroBot"
	"ZeroBot/message"
)

const (
	pluginName = "图包相关"
)

var (
	typeMap = map[string]string{
		"CG":    cgType,
		"cg":    cgType,
		"表情包":   emoticonType,
		"表情":    emoticonType,
		"emoji": emoticonType,
	}
	reTypeStr = `(CG|cg|表情包|表情|emoji)`
	reType    = regexp.MustCompile(reTypeStr)
)

func init() {
	engine := zero.NewTemplate(&zero.Metadata{
		Name: pluginName,
		Help: `-随机xxx[CG|cg|表情包|表情]`,
	})
	db := database.GetDefalutDB()
	db.AutoMigrate(&picturePackage{})
	gdb = (*ymgaldb)(db)

	engine.OnPrefixGroup([]string{"来点", "随机"}).SecondPriority().SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			picType := ""
			if matched := reType.FindStringSubmatch(param); matched != nil {
				picType = typeMap[matched[1]]
				param = reType.ReplaceAllString(param, "")
			}
			key := strings.TrimSpace(param)

			if picType == "" && key == "" {
				return
			}
			y := gdb.getRandPic(picType, key)
			sendYmgal(y, ctx)
		},
	)

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

	engine.OnFullMatch("更新本地图片", zero.SuperUserPermission).
		SetBlock(true).SetNoTimeOut(true).Handle(
		func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			err := updateLocalPic()
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.Send("本地图库关联数据库已更新")
		})
}

func sendYmgal(y picturePackage, ctx *zero.Ctx) {
	if y.PictureList == "" {
		ctx.SendChain(message.Text(zero.BotConfig.NickName[0] + "暂时没有这样的图呢"))
		return
	}

	urlList := strings.Split(y.PictureList, ",")
	url := urlList[rand.IntN(len(urlList))]
	_, err := ctx.SendChain(message.Text(y.Title), message.Image(url))
	if err != nil {
		ctx.SendChain(message.Text("该图发不出..."))
	}
}
