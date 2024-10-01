// Package emojimix 合成emoji
package emojimix

import (
	emoji_map "MiniBot/plugin/emojimix/proto"
	"MiniBot/utils/path"
	"os"
	"path/filepath"
	"strconv"

	zero "ZeroBot"

	"ZeroBot/message"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

var emojiMap = emoji_map.OuterMap{OuterMap: map[int64]*emoji_map.InnerMap{}}
var pluginName = "emojimix"
var dataPath = path.GetPluginDataPath()

// const bed = "https://www.gstatic.com/android/keyboard/emojikitchen/%d/u%x/u%x_u%x.png"

func init() {
	data, err := os.ReadFile(filepath.Join(dataPath, "outer_map.bin"))
	if err != nil {
		log.Error().Err(err).Str("name", pluginName).Msg("")
		return
	}
	if err := proto.Unmarshal(data, &emojiMap); err != nil {
		log.Error().Err(err).Str("name", pluginName)
		return
	}
	zero.NewTemplate(&zero.MetaData{
		Name: "合成emoji",
		Help: "- [emoji][emoji]",
	}).OnMessage(match).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Image(ctx.State["emojimix"].(string)))
		})
}

func match(ctx *zero.Ctx) bool {
	if len(ctx.Event.Message) == 2 {
		r1 := face2emoji(ctx.Event.Message[0])
		r2 := face2emoji(ctx.Event.Message[1])
		if setUrl(r1, r2, ctx) {
			return true
		}
	}

	r := []rune(ctx.Event.RawMessage)
	if len(r) == 2 {
		r1 := int64(r[0])
		r2 := int64(r[1])

		if setUrl(r1, r2, ctx) {
			return true
		}
	}
	return false
}

func setUrl(i int64, j int64, ctx *zero.Ctx) bool {
	if interMap, ok := emojiMap.OuterMap[i]; ok {
		if url, ok := interMap.InnerMap[j]; ok {
			ctx.State["emojimix"] = url
			return true
		}
	}
	return false
}

func face2emoji(face message.MessageSegment) int64 {
	if face.Type == "text" {
		r := []rune(face.Data["text"])
		if len(r) != 1 {
			return 0
		}
		return int64(r[0])
	}
	if face.Type != "face" {
		return 0
	}
	id, err := strconv.Atoi(face.Data["id"])
	if err != nil {
		return 0
	}
	if r, ok := qqface[id]; ok {
		return r
	}
	return 0
}
