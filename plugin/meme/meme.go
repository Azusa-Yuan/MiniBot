package meme

import (
	"MiniBot/utils/cache"
	"MiniBot/utils/net_tools"
	zero "ZeroBot"
	"ZeroBot/message"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

func init() {
	err := InitMeme()
	if err != nil {
		logrus.Error(err)
		return
	}
	metaData := zero.MetaData{
		Name: "表情包制作",
		Help: "发送 表情包列表 查看所有表情指令 \n 发送 查看表情信息xx 查看表情详细参数",
	}
	engine := zero.NewTemplate(&metaData)
	engine.OnFullMatch("表情包列表").SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			data, err := GetHelp()
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprint("[meme]", err)))
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.ImageBytes(data))
		},
	)

	pattern := `--(\S+)\s+(\S+)`
	reArgs := regexp.MustCompile(pattern)
	engine.OnPrefixGroup(maps.Keys(emojiMap)).Handle(
		func(ctx *zero.Ctx) {
			path := emojiMap[ctx.State["prefix"].(string)]
			imgStrs := []string{}
			args := map[string]any{}
			args["user_infos"] = []UserInfo{}

			for i, segment := range ctx.Event.Message {
				if i == 0 {
					continue
				}
				if segment.Type == "at" {
					qqStr := segment.Data["qq"]
					qq, err := strconv.ParseInt(qqStr, 10, 64)
					if err != nil {
						logrus.Errorln("[meme]", err)
						continue
					}
					args["user_infos"] = append(args["user_infos"].([]UserInfo),
						UserInfo{
							Name:   ctx.CardOrNickName(qq),
							Gender: "female",
						})
					imgStrs = append(imgStrs, qqStr)
				}
				if segment.Type == "image" {
					logrus.Debug(segment)
					imgStrs = append(imgStrs, segment.Data["url"])
				}
			}

			// 正则匹配所有args
			extractPlainText := ctx.State["args"].(string)
			matchs := reArgs.FindAllStringSubmatch(extractPlainText, -1)
			if len(matchs) > 0 {
				for _, match := range matchs {
					args[match[1]] = match[2]
				}
				extractPlainText = reArgs.ReplaceAllString(extractPlainText, "")
			}
			// Fields函数会将字符串按空格分割,并自动忽略连续的空格
			texts := strings.Fields(extractPlainText)
			logrus.Debugln(texts)
			logrus.Debugln(imgStrs)

			if !fastJudge(path, len(imgStrs), len(texts)) {
				imgStrs = append([]string{strconv.FormatInt(ctx.Event.UserID, 10)}, imgStrs...)
				args["user_infos"] = append(args["user_infos"].([]UserInfo),
					UserInfo{
						Name:   ctx.CardOrNickName(ctx.Event.UserID),
						Gender: "female",
					})
				if !fastJudge(path, len(imgStrs), len(texts)) {
					return
				}
			}

			// 匹配上了才组织后续
			ctx.Stop()

			images, err := dealImgStr(imgStrs...)
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprint("[meme]", err)))
				return
			}

			argsBytes, err := json.Marshal(args)
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprint("[meme]", err)))
				return
			}

			emojiData, err := CreateEmoji(path, images, texts, string(argsBytes))
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprint("[meme]", err)))
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.ImageBytes(emojiData))
		},
	)

	engine.OnPrefix("查看表情信息").SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			key := strings.TrimSpace(ctx.State["args"].(string))
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(QueryEmojiInfo(key)))
		},
	)
}

func dealImgStr(imgStrs ...string) ([][]byte, error) {
	images := [][]byte{}
	for _, imgStr := range imgStrs {
		uid, err := strconv.ParseInt(imgStr, 10, 64)
		var data []byte
		if err != nil {
			data, err = net_tools.DownloadWithoutTLSVerify(imgStr)
		} else {
			data, err = cache.GetAvatar(uid)
		}
		if err != nil {
			return nil, err
		}
		images = append(images, data)
	}
	return images, nil
}
