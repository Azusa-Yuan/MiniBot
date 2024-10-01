package asill

import (
	"MiniBot/utils/path"
	zero "ZeroBot"
	"encoding/json"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"ZeroBot/message"

	"github.com/rs/zerolog/log"
)

var (
	help = `[发病 对象] 对发病对象发病
[小作文] 随机发送一篇发病小作文
[病情加重 对象/小作文] 将一篇发病小作文添加到数据库中（必须带“/”）
[病情查重 小作文] 对一篇小作文进行查重
`

// [<回复一个小作文> 病情查重] 同上
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
		Name: "asill",
		Help: help,
	})

	engine.OnPrefix(`发病`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		name := ctx.NickName()
		n := rand.IntN(len(data))
		asilldata := data[n]
		asillText := asilldata.Text
		asillText = strings.Replace(asillText, asilldata.Person, name, -1)

		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(asillText))
	})

	// engine.OnPrefix(`病情查重`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
	// 	name := ctx.NickName()
	// 	n := rand.IntN(len(data))
	// 	asilldata := data[n]
	// 	asillText := asilldata.Text
	// 	asillText = strings.Replace(asillText, asilldata.Person, name, -1)

	// 	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(asillText))
	// })
}
