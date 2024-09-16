package genai

import (
	ai "MiniBot/utils/AI"
	"MiniBot/utils/cache"
	"MiniBot/utils/net_tools"
	"MiniBot/utils/transform"
	zero "ZeroBot"
	"ZeroBot/message"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

func init() {
	metaData := zero.MetaData{
		Name: "genai",
		Help: `使用Google大模型Gemini生成,快和露露姆进行ai对话
指令:
	重置会话:可以重置当前ai会话的上下文,默认上下文的时长为2小时
	切换人格[人格]:可以切换不同的对话风格，默认的对话风格为露露姆(虽然出于安全的考虑,已经OOC了),人格为空就是没有人格
				目前支持的人格:露露姆,伊蕾娜,爱莉希雅`,
	}
	engine := zero.NewTemplate(&metaData)
	engine.OnMessage(zero.OnlyToMe).SetPriority(10).Handle(
		func(ctx *zero.Ctx) {
			// 私聊的时候每个qq独立的会话，群聊的时候一个群一个会话
			key := transform.BidWithuidInt64(ctx)
			if ctx.Event.GroupID != 0 {
				key = transform.BidWithgidInt64(ctx)
			}

			parts := []genai.Part{}
			msg := ctx.ExtractPlainText()
			msg = strings.TrimSpace(msg)
			if msg != "" {
				parts = append(parts, genai.Text(msg))
			}

			// 获取图片
			imgStrs := []string{}
			imgTypes := []string{}
			for _, segment := range ctx.Event.Message {
				if segment.Type == "image" {
					imgStrs = append(imgStrs, segment.Data["url"])
					filePath := strings.Split(segment.Data["file"], ".")
					imgTypes = append(imgTypes, filePath[len(filePath)-1])
				}
			}

			imgBytes, err := dealImgStr(imgStrs...)
			if err != nil {
				ctx.SendError(err)
				return
			}

			for i := 0; i < len(imgTypes); i++ {
				if imgTypes[i] == "jpg" {
					imgTypes[i] = "jpeg"
				}
				parts = append(parts, genai.ImageData(imgTypes[i], imgBytes[i]))
			}

			if len(parts) == 0 {
				return
			}
			resp, err := ai.AIBot.SendPartsWithSession(key, parts...)

			if err != nil {
				if err.Error() == "not session" {
					ai.AIBot.CreateSession(key, ai.IM.IntroduceMap["露露姆"])
					resp, err = ai.AIBot.SendPartsWithSession(key, parts...)
				}
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[ai] ", err)))
					return
				}
			}

			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(resp))
		})
	engine.OnFullMatch("重置会话", zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			// 私聊的时候每个qq独立的会话，群聊的时候一个群一个会话
			key := transform.BidWithuidInt64(ctx)
			if ctx.Event.GroupID != 0 {
				key = transform.BidWithgidInt64(ctx)
			}
			ai.AIBot.DelSession(key)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("重置成功"))
		})

	engine.OnPrefix("切换人格", zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			role := strings.TrimSpace(ctx.State["args"].(string))
			if role != "" {
				role = ai.IM.IntroduceMap[role]
				if role == "" {
					ctx.SendChain(message.At(ctx.Event.UserID), message.Text("没有该人格"))
					return
				}
			}

			key := transform.BidWithuidInt64(ctx)
			if ctx.Event.GroupID != 0 {
				key = transform.BidWithgidInt64(ctx)
			}
			ai.AIBot.CreateSession(key, role)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("切换人格成功"))
		})
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
