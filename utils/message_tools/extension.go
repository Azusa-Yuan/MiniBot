package message_tools

import (
	zero "ZeroBot"
	"ZeroBot/message"
)

// FakeSenderForwardNode ...
func FakeSenderForwardNode(ctx *zero.Ctx, msgs ...message.MessageSegment) message.MessageSegment {
	return message.CustomNode(
		ctx.CardOrNickName(ctx.Event.UserID),
		ctx.Event.UserID,
		msgs)
}

// SendFakeForwardToGroup ...
// func SendFakeForwardToGroup(ctx *zero.Ctx, msgs ...message.MessageSegment) NoCtxSendMsg {
// 	return func(msg any) int64 {
// 		return ctx.SendGroupForwardMessage(ctx.Event.GroupID, message.Message{
// 			FakeSenderForwardNode(ctx, msg.(message.Message)...),
// 			FakeSenderForwardNode(ctx, msgs...),
// 		}).Get("message_id").Int()
// 	}
// }
