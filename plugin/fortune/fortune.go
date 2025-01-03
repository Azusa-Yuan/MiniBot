// Package fortune 每日运势
package fortune

import (
	"MiniBot/utils"
	"MiniBot/utils/cache"
	database "MiniBot/utils/db"
	"MiniBot/utils/path"
	"archive/zip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hash/crc64"
	"image"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"unsafe"

	"github.com/FloatTech/imgfactory"

	zero "ZeroBot"

	"ZeroBot/message"

	"github.com/FloatTech/gg" // 注册了 jpg png gif
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/clause"
)

var (
	// 底图缓存位置
	images = filepath.Join(path.GetPluginDataPath())
	// 基础文件位置
	omikujson = filepath.Join(images, "text.json")
	// 生成图缓存位置
	cacheImg   = images + "cache/"
	pluginName = "每日运势"
	// 底图类型列表
	table = [...]string{"车万", "DC4", "爱因斯坦", "星空列车", "樱云之恋", "富婆妹",
		"李清歌", "公主连结", "原神", "明日方舟", "碧蓝航线", "碧蓝幻想",
		"战双", "阴阳师", "赛马娘", "东方归言录", "奇异恩典", "夏日口袋", "ASoul", "Hololive"}
	// 映射底图与 index
	index = make(map[string]uint8)
	// 签文
	omikujis []map[string]string
)

type fortune struct {
	GroupID int64  `gorm:"column:gid; uniqueIndex:gid_bid"`
	BotID   int64  `gorm:"column:bid; uniqueIndex:gid_bid"`
	Value   string `gorm:"column:value"`
}

func init() {
	en := zero.NewTemplate(&zero.Metadata{
		Name: pluginName,
		Help: "- 运势 | 抽签\n" +
			"- 设置底图[车万 | DC4 | 爱因斯坦 | 星空列车 | 樱云之恋 | 富婆妹 | 李清歌 | 公主连结 | 原神 | 明日方舟 | 碧蓝航线 | 碧蓝幻想 | 战双 | 阴阳师 | 赛马娘 | 东方归言录 | 奇异恩典 | 夏日口袋 | ASoul | Hololive]",
	})
	db := database.GetDefalutDB()
	db.AutoMigrate(&fortune{})
	_ = os.RemoveAll(cacheImg)
	err := os.MkdirAll(cacheImg, 0755)
	if err != nil {
		panic(err)
	}
	for i, s := range table {
		index[s] = uint8(i)
	}

	data, err := os.ReadFile(omikujson)
	if err != nil {
		log.Error().Err(err).Str("name", pluginName).Msg("")
		return
	}
	err = json.Unmarshal(data, &omikujis)
	if err != nil {
		log.Error().Err(err).Str("name", pluginName).Msg("")
		return
	}

	en.OnRegex(`^设置底图\s?(.*)`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}
			_, ok := index[ctx.State["regex_matched"].([]string)[1]]
			if ok {
				fortuneInfo := fortune{
					GroupID: gid,
					BotID:   ctx.Event.SelfID,
					Value:   ctx.State["regex_matched"].([]string)[1],
				}
				// 如果发生冲突则什么都不做
				if row := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&fortuneInfo).RowsAffected; row == 0 {
					err := db.Model(&fortune{}).Where("gid = ? AND bid = ?", gid, ctx.Event.SelfID).
						Update("value", fortuneInfo.Value).Error
					if err != nil {
						ctx.SendError(err)
						return
					}
				}

				ctx.SendChain(message.Text("设置成功~"))
				return
			}
			ctx.SendChain(message.Text("没有这个底图哦～"))
		})

	en.OnFullMatchGroup([]string{"运势", "抽签", "今日运势"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {

			kind := "碧蓝航线"
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}

			fortuneInfo := fortune{}
			db.Where("gid = ? AND bid = ?", gid, ctx.Event.SelfID).First(&fortuneInfo)
			if fortuneInfo.Value != "" {
				kind = fortuneInfo.Value
			}

			zipfile := filepath.Join(images, kind+".zip")

			// 随机获取背景
			background, index, err := randimage(zipfile, ctx)
			if err != nil {
				ctx.SendError(err)
				return
			}
			// 随机获取签文
			randtextindex := RandSenderPerDayN(ctx.Event.UserID, len(omikujis))
			title, text := omikujis[randtextindex]["title"], omikujis[randtextindex]["content"]
			digest := md5.Sum(utils.StringToBytes(zipfile + strconv.Itoa(index) + title + text))
			cachefile := cacheImg + hex.EncodeToString(digest[:])

			imgStr, _ := cache.GetRedisCli().Get(context.Background(), cachefile).Result()
			if imgStr != "" {
				ctx.SendChain(message.ImageBytes(utils.StringToBytes(imgStr)))
				return
			}

			fontdata, _ := cache.GetDefaultFont()
			imgData, err := draw(background, fontdata, title, text)
			if err != nil {
				ctx.SendError(err)
				return
			}

			cache.GetRedisCli().Set(context.Background(), cachefile, utils.BytesToString(imgData), 24*time.Hour)

			ctx.SendChain(message.ImageBytes(imgData))
		})
}

// @function randimage 随机选取zip内的文件
// @param path zip路径
// @param ctx *zero.Ctx
// @return 文件路径 & 错误信息
func randimage(path string, ctx *zero.Ctx) (im image.Image, index int, err error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	defer reader.Close()

	file := reader.File[RandSenderPerDayN(ctx.Event.UserID, len(reader.File))]
	f, err := file.Open()
	if err != nil {
		return
	}
	defer f.Close()

	im, _, err = image.Decode(f)
	return
}

// @function draw 绘制运势图
// @param background 背景图片路径
// @param seed 随机数种子
// @param title 签名
// @param text 签文
// @return 错误信息
func draw(back image.Image, fontdata []byte, title, txt string) ([]byte, error) {
	canvas := gg.NewContext(back.Bounds().Size().Y, back.Bounds().Size().X)
	canvas.DrawImage(back, 0, 0)
	// 写标题
	canvas.SetRGB(1, 1, 1)
	if err := canvas.ParseFontFace(fontdata, 45); err != nil {
		return nil, err
	}
	sw, _ := canvas.MeasureString(title)
	canvas.DrawString(title, 140-sw/2, 112)
	// 写正文
	canvas.SetRGB(0, 0, 0)
	if err := canvas.ParseFontFace(fontdata, 23); err != nil {
		return nil, err
	}
	tw, th := canvas.MeasureString("测")
	tw, th = tw+10, th+10
	r := []rune(txt)
	xsum := rowsnum(len(r), 9)
	switch xsum {
	default:
		for i, o := range r {
			xnow := rowsnum(i+1, 9)
			ysum := min(len(r)-(xnow-1)*9, 9)
			ynow := i%9 + 1
			canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(ysum, ynow, th)+320.0)
		}
	case 2:
		div := rowsnum(len(r), 2)
		for i, o := range r {
			xnow := rowsnum(i+1, div)
			ysum := min(len(r)-(xnow-1)*div, div)
			ynow := i%div + 1
			switch xnow {
			case 1:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow, th)+320.0)
			case 2:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow+(9-ysum), th)+320.0)
			}
		}
	}
	return imgfactory.ToBytes(canvas.Image())
}

func offest(total, now int, distance float64) float64 {
	if total%2 == 0 {
		return (float64(now-total/2) - 1) * distance
	}
	return (float64(now-total/2) - 1.5) * distance
}

func rowsnum(total, div int) int {
	temp := total / div
	if total%div != 0 {
		temp++
	}
	return temp
}

// RandSenderPerDayN 每个用户每天随机数
func RandSenderPerDayN(uid int64, n int) int {
	sum := crc64.New(crc64.MakeTable(crc64.ISO))
	sum.Write(utils.StringToBytes(time.Now().Format("20060102")))
	sum.Write((*[8]byte)(unsafe.Pointer(&uid))[:])
	r := rand.New(rand.NewSource(int64(sum.Sum64())))
	return r.Intn(n)
}
