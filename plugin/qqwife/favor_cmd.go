package qqwife

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"

	zero "ZeroBot"
	"ZeroBot/message"

	"MiniBot/service/wallet"
	"MiniBot/utils/cache"

	"github.com/FloatTech/imgfactory"

	"github.com/FloatTech/gg"
)

func init() {
	// 好感度系统要用到货币系统  所以单独出来
	engine.OnRegex(`^GetFavorability\s*(\[CQ:at,qq=)?(\d+)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			fiancee, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			uid := ctx.Event.UserID
			favor, err := qqwife.GetFavorability(uid, fiancee)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			// 输出结果
			ctx.SendChain(
				message.At(uid),
				message.Text("\n当前你们好感度为", favor),
			)
		})
	// 礼物系统
	engine.OnPrefix("买礼物给").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			atInfos := ctx.GetAtInfos()
			if len(atInfos) != 1 {
				return
			}
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID

			gay := atInfos[0].QQ
			if gay == uid {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.At(uid), message.Text("你想给自己买什么礼物呢?")))
				return
			}
			// 获取CD
			groupInfo, err := qqwife.GetGroupInfo(gid)
			if err != nil {
				ctx.SendError(err)
				return
			}
			timeMin, err := qqwife.JudgeCD(gid, uid, "买礼物", groupInfo.CDtime)
			if err != nil {
				ctx.SendError(err)
				return
			}

			if timeMin > 0 {
				ctx.SendChain(message.Text(fmt.Sprintf("舔狗，今天你已经送过礼物了。等%d分钟再舔吧", timeMin)))
				return
			}
			// 获取好感度
			favor, err := qqwife.GetFavorability(uid, gay)
			if err != nil {
				ctx.SendError(err)
				return
			}
			// 对接小熊饼干
			money := wallet.GetWalletMoneyByCtx(ctx)
			if money < 1 {
				ctx.SendChain(message.Text("你钱包没钱啦！"))
				return
			}
			moneyToFavor := rand.IntN(min(money, 100)) + 1
			// 计算钱对应的好感值
			newFavor := 1
			moodMax := 3
			if favor > 50 {
				newFavor = moneyToFavor % 10 // 礼物厌倦
			} else {
				moodMax = 5
				newFavor += rand.IntN(moneyToFavor)
			}
			// 随机对方心情 好感度大于50就是50%喜欢 50%讨厌了
			mood := rand.IntN(moodMax)
			if mood == 0 {
				newFavor = -newFavor
			}
			// 记录结果
			err = wallet.UpdateWalletByCtx(ctx, -moneyToFavor)
			if err != nil {
				ctx.SendError(err)
				return
			}
			lastfavor, err := qqwife.UpdateFavorability(uid, gay, newFavor)
			if err != nil {
				ctx.SendError(err)
				return
			}
			// 写入CD
			err = qqwife.SaveCD(gid, uid, "买礼物")
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[ERROR]:你的技能CD记录失败\n", err))
			}
			// 输出结果
			if mood == 0 {
				ctx.SendChain(message.Text("你花了", moneyToFavor, wallet.GetWalletName(), "买了一件女装送给了ta,ta很不喜欢,你们的好感度降低至", lastfavor))
			} else {
				ctx.SendChain(message.Text("你花了", moneyToFavor, wallet.GetWalletName(), "买了一件女装送给了ta,ta很喜欢,你们的好感度升至", lastfavor))
			}
		})
	engine.OnFullMatch("好感度列表", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			infos, err := qqwife.GetFavorabilityList(uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
				return
			}
			// 限制名单数量
			sort.Slice(infos, func(i, j int) bool {
				return infos[i].Favor > infos[j].Favor // 降序排序
			})
			number := len(infos)
			if number > 10 {
				number = 10
				infos = infos[:10]
			}
			/***********设置图片的大小和底色***********/

			fontSize := 50.0
			canvas := gg.NewContext(1150, int(170+(50+70)*float64(number)))
			canvas.SetRGB(1, 1, 1) // 白色
			canvas.Clear()
			/***********下载字体***********/
			data, err := cache.GetDefaultFont()
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(data, fontSize*2); err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
				return
			}
			sl, h := canvas.MeasureString("你的好感度排行列表")
			/***********绘制标题***********/
			canvas.DrawString("你的好感度排行列表", (1100-sl)/2, 100) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 160)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(data, fontSize); err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
				return
			}
			for i := 0; i < number; i++ {
				info := infos[i]
				target := info.Target
				userName := ctx.CardOrNickName(target)
				canvas.SetRGB255(0, 0, 0)
				canvas.DrawString(userName+fmt.Sprintf("(%d)", target), 10, float64(180+(50+70)*i))
				canvas.DrawString(strconv.Itoa(info.Favor), 1020, float64(180+60+(50+70)*i))
				canvas.DrawRectangle(10, float64(180+60+(50+70)*i)-h/2, 1000, 50)
				canvas.SetRGB255(150, 150, 150)
				canvas.Fill()
				canvas.SetRGB255(0, 0, 0)
				canvas.DrawRectangle(10, float64(180+60+(50+70)*i)-h/2, float64(info.Favor)*10, 50)
				canvas.SetRGB255(231, 27, 100)
				canvas.Fill()
			}
			data, err = imgfactory.ToBytes(canvas.Image())
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	// engine.OnFullMatch("好感度数据整理", zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByUser).
	// 	Handle(func(ctx *zero.Ctx) {
	// 		ctx.SendChain(message.Text("开始整理力，请稍等"))
	// 		qqwife.Lock()
	// 		defer qqwife.Unlock()
	// 		var count int64
	// 		res := qqwife.db.Model(&favorability{}).Count(&count)
	// 		if res.Error != nil {
	// 			ctx.SendChain(message.Text("[ERROR]: ", res.Error))
	// 			return
	// 		}
	// 		if count == 0 {
	// 			ctx.SendChain(message.Text("[ERROR]: 不存在好感度数据."))
	// 			return
	// 		}
	// 		favor := favorability{}
	// 		delInfo := make([]string, 0, count*2)
	// 		favorInfo := make(map[string]int, count*2)
	// 		_ = qqwife.db.FindFor("favorability", &favor, "group by Userinfo", func() error {
	// 			delInfo = append(delInfo, favor.Userinfo)
	// 			// 解析旧数据
	// 			userList := strings.Split(favor.Userinfo, "+")
	// 			maxQQ, _ := strconv.ParseInt(userList[0], 10, 64)
	// 			minQQ, _ := strconv.ParseInt(userList[1], 10, 64)
	// 			if maxQQ > minQQ {
	// 				favor.Userinfo = userList[0] + "+" + userList[1]
	// 			} else {
	// 				favor.Userinfo = userList[1] + "+" + userList[0]
	// 			}
	// 			// 判断是否是重复的
	// 			score, ok := favorInfo[favor.Userinfo]
	// 			if ok {
	// 				if score < favor.Favor {
	// 					favorInfo[favor.Userinfo] = favor.Favor
	// 				}
	// 			} else {
	// 				favorInfo[favor.Userinfo] = favor.Favor
	// 			}
	// 			return nil
	// 		})
	// 		for _, updateinfo := range delInfo {
	// 			// 删除旧数据
	// 			err = qqwife.db.Del("favorability", "where Userinfo = '"+updateinfo+"'")
	// 			if err != nil {
	// 				userList := strings.Split(favor.Userinfo, "+")
	// 				uid1, _ := strconv.ParseInt(userList[0], 10, 64)
	// 				uid2, _ := strconv.ParseInt(userList[1], 10, 64)
	// 				ctx.SendChain(message.Text("[ERROR]: 删除", ctx.CardOrNickName(uid1), "和", ctx.CardOrNickName(uid2), "的好感度时发生了错误。\n错误信息:", err))
	// 			}
	// 		}
	// 		for userInfo, favor := range favorInfo {
	// 			favorInfo := favorability{
	// 				Userinfo: userInfo,
	// 				Favor:    favor,
	// 			}
	// 			err = qqwife.db.Insert("favorability", &favorInfo)
	// 			if err != nil {
	// 				userList := strings.Split(userInfo, "+")
	// 				uid1, _ := strconv.ParseInt(userList[0], 10, 64)
	// 				uid2, _ := strconv.ParseInt(userList[1], 10, 64)
	// 				ctx.SendChain(message.Text("[ERROR]: 更新", ctx.CardOrNickName(uid1), "和", ctx.CardOrNickName(uid2), "的好感度时发生了错误。\n错误信息:", err))
	// 			}
	// 		}
	// 		ctx.SendChain(message.Text("清理好了哦"))
	// 	})
}
