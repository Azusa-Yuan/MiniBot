package genai

import (
	ai "MiniBot/utils/AI"
	"MiniBot/utils/transform"
	zero "ZeroBot"
	"ZeroBot/message"
	"fmt"
	"strings"
)

func init() {
	metaData := zero.MetaData{
		Name: "genai",
		Help: "使用google api生成",
	}
	engine := zero.NewTemplate(&metaData)
	engine.OnMessage(zero.OnlyToMe).SetPriority(10).Handle(
		func(ctx *zero.Ctx) {
			// 私聊的时候每个qq独立的会话，群聊的时候一个群一个会话
			key := transform.BidWithuidInt64(ctx)
			if ctx.Event.GroupID != 0 {
				key = transform.BidWithgidInt64(ctx)
			}
			resp, err := ai.AIBot.SendMsgWithSession(key, ctx.ExtractPlainText())
			if err != nil {
				if err.Error() == "not session" {
					ai.AIBot.CreateSession(key, ai.IM.IntroduceMap["露露姆"])
					resp, err = ai.AIBot.SendMsgWithSession(key, ctx.ExtractPlainText())
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
