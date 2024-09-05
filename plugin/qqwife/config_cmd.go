// Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品，文案采用了Hana的zbp娶群友文案
package qqwife

import (
	"strconv"

	zero "ZeroBot"
	"ZeroBot/message"
	// 反并发
	// 数据库
	// Sqlite driver based on CGO
	// "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	// 画图
)

// 使用 var() 和直接声明多个全局变量之间的主要区别在于代码的可读性和组织性。然而，对于小型程序或者变量数量不多的情况，直接声明全局变量也是可以接受的，具体取决于个人或团队的偏好和项目的需求。

func init() {
	// engine.OnRegex(`^重置(所有|本群|/d+)?花名册$`, zero.SuperUserPermission).SetBlock(true).Limit(limiter.LimitByGroup).
	// 	Handle(func(ctx *zero.Ctx) {
	// 		var err error
	// 		switch ctx.State["regex_matched"].([]string)[1] {
	// 		case "所有":
	// 			err = qqwife.清理花名册()
	// 		case "本群", "":
	// 			if ctx.Event.GroupID == 0 {
	// 				ctx.SendChain(message.Text("该功能只能在群组使用或者指定群组"))
	// 				return
	// 			}
	// 			err = qqwife.清理花名册("group" + strconv.FormatInt(ctx.Event.GroupID, 10))
	// 		default:
	// 			cmd := ctx.State["regex_matched"].([]string)[1]
	// 			gid, _ := strconv.ParseInt(cmd, 10, 64) // 判断是否为群号
	// 			if gid == 0 {
	// 				ctx.SendChain(message.Text("请输入正确的群号"))
	// 				return
	// 			}
	// 			err = qqwife.清理花名册("group" + cmd)
	// 		}
	// 		if err != nil {
	// 			ctx.SendChain(message.Text("[ERROR]:", err))
	// 			return
	// 		}
	// 		ctx.SendChain(message.Text("重置成功"))
	// 	})
	engine.OnRegex(`^设置CD为(\d+)小时`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			cdTime, err := strconv.ParseFloat(ctx.State["regex_matched"].([]string)[1], 64)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]请设置纯数字\n", err))
				return
			}
			groupInfo, err := qqwife.GetGroupInfo(ctx.Event.GroupID)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			groupInfo.CDtime = cdTime
			err = qqwife.UpdateGroupInfo(groupInfo)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]设置CD时长失败\n", err))
				return
			}
			ctx.SendChain(message.Text("设置成功"))
		})
	engine.OnRegex(`^(允许|禁止)(自由恋爱|牛头人)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			status := ctx.State["regex_matched"].([]string)[1]
			mode := ctx.State["regex_matched"].([]string)[2]
			groupInfo, err := qqwife.GetGroupInfo(ctx.Event.GroupID)
			switch {
			case err != nil:
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			case mode == "自由恋爱":
				if status == "允许" {
					groupInfo.CanMatch = 1
				} else {
					groupInfo.CanMatch = 0
				}
			case mode == "牛头人":
				if status == "允许" {
					groupInfo.CanNtr = 1
				} else {
					groupInfo.CanNtr = 0
				}
			}
			err = qqwife.UpdateGroupInfo(groupInfo)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Text("设置成功"))
		})

}
