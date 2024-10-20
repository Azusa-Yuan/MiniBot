// 部分参考 Package picturePackage 月幕galgame
package picturepackage

import (
	database "MiniBot/utils/db"
	"MiniBot/utils/net_tools"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	zero "ZeroBot"
	"ZeroBot/message"

	"github.com/tidwall/gjson"
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
		Help: `-随机xxx[CG|cg|表情包|表情]
随机xxx[CG|cg|表情包|表情]`,
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
			sendYmgal(y, ctx, key, picType)
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

func sendYmgal(y picturePackage, ctx *zero.Ctx, key, picType string) {
	if y.PictureList == "" {
		if picType != emoticonType {
			encodedTag := url.QueryEscape("size=regular&tag=" + key)
			resp, err := http.DefaultClient.Get("https://api.lolicon.app/setu/v2?" + encodedTag)
			if err == nil && resp.StatusCode == http.StatusOK {
				respData, err := io.ReadAll(resp.Body)
				if err == nil {
					dataArray := gjson.ParseBytes(respData).Get("data").Array()
					if len(dataArray) != 0 {
						imgData := dataArray[0]
						url := imgData.Get("urls").Get("original").String()
						y.PictureList = url
						y.Title = imgData.Get("title").String() + "/" + imgData.Get("author").String()
						y.Title = zero.BotConfig.NickName[0] + "暂时没有这样的图呢。" + "所以给你发这张:" + y.Title
					}
				}
			}
		}
		if y.PictureList == "" {
			ctx.SendChain(message.Text(zero.BotConfig.NickName[0] + "暂时没有这样的图呢"))
			return
		}
	}

	// napcat 下载慢 不明原因
	urlList := strings.Split(y.PictureList, ",")
	url := urlList[rand.IntN(len(urlList))]
	var err error
	var imgBytes []byte
	if strings.HasPrefix(url, "http") {
		imgBytes, err = net_tools.DownloadWithoutTLSVerify(url)
		if err == nil {
			_, err = ctx.SendChain(message.Text(y.Title), message.ImageBytes(imgBytes))
		}
	} else {
		_, err = ctx.SendChain(message.Text(y.Title), message.Image(url))
	}
	if err != nil {
		ctx.SendError(fmt.Errorf("该图发不出。。。"), message.Text(url))
	}
}
