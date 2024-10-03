package asill

import (
	"MiniBot/utils/path"
	zero "ZeroBot"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"ZeroBot/message"

	"github.com/rs/zerolog/log"
	"github.com/xrash/smetrics"
)

var (
	help = `[发病 对象] 对发病对象发病
[小作文] 随机发送一篇发病小作文
[病情加重 对象/小作文] 将一篇发病小作文添加到数据库中（必须带“/”）
[病情查重 小作文] 对一篇小作文进行查重
`

	// [<回复一个小作文> 病情查重] 同上
	pluginName = "asill"
)

type asillData struct {
	Person string `json:"person"`
	Text   string `json:"text"`
}

func init() {
	dataPath := filepath.Join(path.GetDataPath(), "data.json")
	data := []asillData{}

	rawdata, err := os.ReadFile(dataPath)
	if err != nil {
		log.Error().Msgf("[asill] Error reading file: %v, not load this plugin", err)
		return
	}

	err = json.Unmarshal(rawdata, &data)
	if err != nil {
		log.Error().Msgf("[asill] Error umarshal json: %v, not load this plugin", err)
		return
	}

	engine := zero.NewTemplate(&zero.MetaData{
		Name: pluginName,
		Help: help,
	})

	engine.OnPrefix("发病").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		name := ctx.NickName()
		n := rand.IntN(len(data))
		asilldata := data[n]
		asillText := asilldata.Text
		asillText = strings.Replace(asillText, asilldata.Person, name, -1)

		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(asillText))
	})

	engine.OnPrefix("病情查重").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		msg := ctx.State["args"].(string)
		if ctx.Event.Message[0].Type == "reply" {
			replyMsg := ctx.GetMessage(ctx.Event.Message[0].Data["id"])
			msg = replyMsg.Elements.ExtractPlainText()
		}

		for _, unit := range data {
			rate := smetrics.Jaro(unit.Text, msg)
			if rate > 0.7 {
				ctx.SendChain(message.Text(fmt.Sprintf("在本文库中，总文字复制比：%f%% \n相似小作文：\n%s", 100*rate, unit.Text)))
				return
			}
		}
		ctx.SendChain(message.Text("文库内没有相似的小作文"))
	})

	engine.OnPrefix("病情加重", zero.SuperUserPermission).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			msg := ctx.State["args"].(string)
			msgs := strings.Split(msg, "/")
			if len(msgs) != 2 {
				ctx.SendChain(message.Text("请发送[病情加重 对象/小作文]（必须带“/”）~"))
				return
			}
			keyWord := msgs[0]
			data = append(data, asillData{
				Person: keyWord,
				Text:   msgs[1],
			})
			rawdata, err = json.MarshalIndent(data, "", " ")
			if err != nil {
				ctx.SendError(err)
				return
			}
			err := os.WriteFile(dataPath, rawdata, 0555)
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.Text("病情已添加"))
		},
	)
}
