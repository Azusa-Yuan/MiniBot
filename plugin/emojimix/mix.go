// Package emojimix 合成emoji
package emojimix

import (
	"MiniBot/utils/path"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	zero "ZeroBot"

	"ZeroBot/message"

	emoji "github.com/Andrew-M-C/go.emoji"
	"github.com/rs/zerolog/log"
)

var emojiMap = map[string]string{}
var pluginName = "emojimix"
var dataPath = path.GetPluginDataPath()

func init() {
	data, err := os.ReadFile(filepath.Join(dataPath, "key_map.json"))
	if err != nil {
		log.Error().Err(err).Str("name", pluginName).Msg("")
		return
	}
	if err := json.Unmarshal(data, &emojiMap); err != nil {
		log.Error().Err(err).Str("name", pluginName)
		return
	}
	zero.NewTemplate(&zero.Metadata{
		Name: "合成emoji",
		Help: "- [emoji][emoji]",
	}).OnMessage(match).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Image(base_url + ctx.State["emojimix"].(string)))
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

	emojis := []string{}
	i := 0
	emoji.ReplaceAllEmojiFunc(ctx.Event.RawMessage, func(emoji string) string {
		if i < 2 {
			emojis = append(emojis, emoji)
		}
		i++
		return ""
	})
	if i == 2 {
		if setUrl(emojis[0], emojis[1], ctx) {
			return true
		}
	}

	return false
}

func setUrl(i string, j string, ctx *zero.Ctx) bool {
	url := getValue(emojiMap, i, j)
	if url != "" {
		ctx.State["emojimix"] = url
		return true
	}
	return false
}

func face2emoji(face message.MessageSegment) string {
	if face.Type == "text" {
		singleEmoji := ""
		i := 0
		emoji.ReplaceAllEmojiFunc(face.Data["text"], func(emoji string) string {
			if i < 1 {
				singleEmoji = emoji
			}
			i++
			return ""
		})
		if i == 1 {
			return singleEmoji
		}
		return ""
	}
	if face.Type != "face" {
		return ""
	}
	id, err := strconv.Atoi(face.Data["id"])
	if err != nil {
		return ""
	}
	if r, ok := QQfaceString[id]; ok {
		return r
	}
	return ""
}
