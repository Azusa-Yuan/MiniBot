// 好友申请以及群聊邀请事件管理处理
package manager

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	zero "ZeroBot"
	"ZeroBot/message"

	"github.com/rs/zerolog/log"
)

func init() {
	// metaData := &zero.MetaData{
	// 	Name: "",
	// 	Help: ,
	// }
	// engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
	// 	DisableOnDefault: false,
	// 	Brief:            "好友申请和群聊邀请事件处理",
	// 	Help: "- [开启|关闭]自动同意[申请|邀请|主人]\n" +
	// 		"- [同意|拒绝][申请|邀请][flag]\n" +
	// 		"Tips: 信息默认发送给主人列表第一位, 默认同意所有主人的事件, flag跟随事件一起发送",
	// })
	zero.On("request/group/invite").SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			userid := ctx.Event.UserID
			username := ctx.CardOrNickName(userid)
			groupname := ctx.GetThisGroupInfo(true).Name
			groupid := ctx.Event.GroupID
			log.Info().Str("name", pluginName).Msgf("收到来自[%s](%d)的群聊邀请，群:[%s](%d)", username, userid, groupname, groupid)

			if zero.SuperUserPermission(ctx) {
				ctx.SetGroupAddRequest(ctx.Event.Flag, "invite", true, "")
				ctx.SendPrivateForwardMessage(su, message.Message{message.CustomNode(username, userid,
					"已自动同意在"+now+"收到来自"+
						"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")的群聊邀请"+
						"\n群聊:["+groupname+"]("+strconv.FormatInt(groupid, 10)+")"+
						"\nflag:"+ctx.Event.Flag)})
				return
			}
			ctx.SendPrivateForwardMessage(su,
				message.Message{message.CustomNode(username, userid,
					"在"+now+"收到来自"+
						"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")的群聊邀请"+
						"\n群聊:["+groupname+"]("+strconv.FormatInt(groupid, 10)+")"+
						"\n请在下方复制flag并在前面加上:"+
						"\n同意/拒绝邀请，来决定同意还是拒绝"),
					message.CustomNode(username, userid, ctx.Event.Flag)})
		})
	zero.On("request/friend").SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			comment := ctx.Event.Comment
			userid := ctx.Event.UserID
			username := ctx.CardOrNickName(userid)
			log.Info().Str("name", pluginName).Msgf("收到来自[%s](%d)的好友申请", username, userid)
			if zero.SuperUserPermission(ctx) {
				ctx.SetFriendAddRequest(ctx.Event.Flag, true, "")
				ctx.SendPrivateMessage(su, message.Text("已自动同意在"+now+"收到来自"+
					"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")"+
					"\n的好友请求:"+comment+
					"\nflag:"+ctx.Event.Flag))
				return
			}
			ctx.SendPrivateMessage(su, message.Text("在"+now+"收到来自"+
				"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")"+
				"\n的好友请求:"+comment+
				"\n请在下方复制flag并在前面加上:"+
				"\n同意/拒绝申请，来决定同意还是拒绝"+"\nflag:"+ctx.Event.Flag))
		})
	zero.OnRegex(`^(同意|拒绝)(申请|邀请)\s*(\S+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			cmd := ctx.State["regex_matched"].([]string)[1]
			org := ctx.State["regex_matched"].([]string)[2]
			flag := ctx.State["regex_matched"].([]string)[3]
			other := ctx.State["regex_matched"].([]string)[4]

			ok := cmd == "同意"
			switch org {
			case "申请":
				ctx.SetFriendAddRequest(flag, ok, other)
				ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
			case "邀请":
				ctx.SetGroupAddRequest(flag, "invite", ok, other)
				ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
			}
		})

	// 退群提醒
	zero.OnNotice().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_decrease" {
				userid := ctx.Event.UserID
				ctx.SendChain(message.Text(ctx.CardOrNickName(userid), "(", userid, ")", "离开了我们..."))
			}
		})

	// // 入群欢迎
	// zero.OnNotice().SetBlock(false).
	// 	Handle(func(ctx *zero.Ctx) {
	// 		if ctx.Event.NoticeType == "group_increase" && ctx.Event.SelfID != ctx.Event.UserID {
	// 		}
	// 	})
	// 非超级权限拉进群
	zero.OnNotice().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_increase" && ctx.Event.SelfID == ctx.Event.UserID {
				if slices.Index(zero.BotConfig.SuperUsers, ctx.Event.OperatorID) >= 0 {
					return
				}
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0],
					fmt.Sprintf("%d 试图将bot拉进群聊 %d", ctx.Event.OperatorID, ctx.Event.GroupID))
				ctx.SendChain(message.Text("不要随意加露露姆进群！"))
				ctx.SetThisGroupLeave(false)
			}
		})

	// zero.OnRegex(`^(开启|关闭)自动同意(申请|邀请|主人)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
	// 	Handle(func(ctx *zero.Ctx) {
	// 		su := zero.BotConfig.SuperUsers[0]
	// 		option := ctx.State["regex_matched"].([]string)[1]
	// 		from := ctx.State["regex_matched"].([]string)[2]
	// 		switch from {
	// 		case "申请":
	// 			data.setapply(option == "开启")
	// 		case "邀请":
	// 			data.setinvite(option == "开启")
	// 		case "主人":
	// 			data.setmaster(option == "关闭")
	// 		}

	// 		ctx.SendChain(message.Text("已设置自动同意" + from + "为" + option))
	// 	})
}
