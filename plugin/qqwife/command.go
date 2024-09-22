// Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品，文案采用了Hana的zbp娶群友文案
package qqwife

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"math/rand/v2"
	"strconv"

	"MiniBot/utils/cache"
	zero "ZeroBot"
	"ZeroBot/message"

	"github.com/FloatTech/gg"
)

var (
	metaData = zero.MetaData{
		Name: "qqwife",
		Help: "- 娶群友\n- 群老婆列表\n- [允许|禁止]自由恋爱\n- [允许|禁止]牛头人\n- 设置CD为xx小时    →(默认1小时)\n- 重置花名册\n- 重置所有花名册(用于清除所有群数据及其设置)\n- GetFavorability[对方Q号|@对方QQ]\n- 好感度列表\n- 好感度数据整理 (当好感度列表出现重复名字时使用)\n" +
			"--------------------------------\n以下指令存在CD,不跨天刷新,前两个受指令开关\n--------------------------------\n" +
			"- (娶|嫁)@对方QQ\n自由选择对象, 自由恋爱(好感度越高成功率越高,保底30%概率)\n" +
			"- 当[对方Q号|@对方QQ]的小三\n我和你才是真爱, 为了你我愿意付出一切(好感度越高成功率越高,保底10%概率)\n" +
			"- 闹离婚\n你谁啊, 给我滚(好感度越高成功率越低)\n" +
			"- 买礼物给[对方Q号|@对方QQ]\n使用小熊饼干获取好感度\n" +
			"- 做媒 @攻方QQ @受方QQ\n身为管理, 群友的xing福是要搭把手的(攻受双方好感度越高成功率越高,保底30%概率)\n" +
			"--------------------------------\n好感度规则\n--------------------------------\n" +
			"\"娶群友\"指令好感度随机增加1~5。\n\"A牛B的C\"会导致C恨A, 好感度-5;\nB为了报复A, 好感度+5(什么柜子play)\nA为BC做媒,成功B、C对A好感度+1反之-1\n做媒成功BC好感度+1" +
			"\nTips: 群老婆列表过0点刷新",
	}
	engine = zero.NewTemplate(&metaData)
)

func init() {
	engine.OnFullMatchGroup([]string{"娶群友", "今日老婆"}, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			err := qqwife.IfToday()
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			uid := ctx.Event.UserID
			userInfo, _ := qqwife.GetMarriageInfo(gid, uid)
			switch {
			case userInfo != (UserInfo{}) && (userInfo.Target == 0 || userInfo.UID == 0): // 如果是单身贵族
				ctx.SendChain(message.Text("今天你是单身贵族噢"))
				return
			case userInfo.Mode == 1: // 娶过别人
				avatarData, err := cache.GetAvatar(userInfo.Target)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
					return
				}
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你在", userInfo.Updatetime.Format("15:04:05"), "娶了群友"),
					message.ImageBytes(avatarData),
					message.Text(
						"\n",
						"[", userInfo.Targetname, "]",
						"(", userInfo.Target, ")哒",
					),
				)
				return
			case userInfo != (UserInfo{}) && userInfo.Mode == 0: // 嫁给别人
				avatarData, err := cache.GetAvatar(userInfo.Target)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
					return
				}
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你在", userInfo.Updatetime.Format("15:04:05"), "被群友"),
					message.ImageBytes(avatarData),
					message.Text(
						"\n",
						"[", userInfo.Targetname, "]",
						"(", userInfo.Target, ")娶了",
					),
				)
				return
			}
			// 有缓存获取群员列表
			botId := ctx.Event.SelfID
			temp, err := cache.GetGroupMemberList(botId, ctx.Event.GroupID)
			if err != nil {
				ctx.SendError(err)
				return
			}

			temp = temp[len(temp)/3:]

			qqwife.Lock()
			defer qqwife.Unlock()
			marriedInfo, err := qqwife.GetAllInfo(gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			// 转换切片为 map，用于快速查找
			marriedMap := make(map[int64]struct{})
			for _, v := range marriedInfo {
				marriedMap[v.UID] = struct{}{}
				// 多一次校验
				if uid == v.UID {
					return
				}
			}
			qqgrouplist := make([]int64, 0, len(temp))
			for k := 0; k < len(temp); k++ {
				uid := temp[k]
				if uid == botId {
					continue
				}
				if _, ok := marriedMap[uid]; ok {
					continue
				}
				qqgrouplist = append(qqgrouplist, uid)
			}
			// 没有人（只剩自己）的时候
			if len(qqgrouplist) == 1 {
				ctx.SendChain(message.Text("~群里没有ta人是单身了哦 明天再试试叭"))
				return
			}
			// 随机抽娶
			fiancee := qqgrouplist[rand.IntN(len(qqgrouplist))]
			if fiancee == uid { // 如果是自己
				err := qqwife.SaveMarriageInfo(gid, uid, 0, "", "")
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Text("今日获得成就：单身贵族"))
			}
			// 去民政局办证
			err = qqwife.SaveMarriageInfo(gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(fiancee))
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}

			// 保存完就用协程去执行，减少锁的时间
			go func() {
				favor, err := qqwife.UpdateFavorability(uid, fiancee, 1+rand.IntN(5))
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
				}
				avatarData, err := cache.GetAvatar(fiancee)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
					return
				}
				// 请大家吃席
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你的群老婆是"),
					message.ImageBytes(avatarData),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(fiancee), "]",
						"(", fiancee, ")哒\n当前你们好感度为", favor,
					),
				)
			}()

		})
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			err := qqwife.IfToday()
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			listAll, err := qqwife.GetAllInfo(gid)
			list := []UserInfo{}
			for _, v := range listAll {
				if v.Mode == 0 {
					continue
				}
				list = append(list, v)
			}
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			number := len(list)
			if number <= 0 {
				ctx.SendChain(message.Text("今天还没有人结婚哦"))
				return
			}
			/***********设置图片的大小和底色***********/
			fontSize := 50.0
			if number < 10 {
				number = 10
			}
			canvas := gg.NewContext(1500, int(250+fontSize*float64(number)))
			canvas.SetRGB(1, 1, 1) // 白色
			canvas.Clear()
			/***********获取字体，可以注销掉***********/
			data, err := cache.GetDefaultFont()
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(data, fontSize*2); err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			sl, h := canvas.MeasureString("群老婆列表")
			/***********绘制标题***********/
			canvas.DrawString("群老婆列表", (1500-sl)/2, 160-h) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 250-h)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(data, fontSize); err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			_, h = canvas.MeasureString("焯")
			for i, info := range list {
				canvas.DrawString(slicename(info.Username, canvas), 0, float64(260+50*i)-h)
				canvas.DrawString("("+strconv.FormatInt(info.UID, 10)+")", 350, float64(260+50*i)-h)
				canvas.DrawString("←→", 700, float64(260+50*i)-h)
				canvas.DrawString(slicename(info.Targetname, canvas), 800, float64(260+50*i)-h)
				canvas.DrawString("("+strconv.FormatInt(info.Target, 10)+")", 1150, float64(260+50*i)-h)
			}
			buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024*4))
			err = jpeg.Encode(buffer, canvas.Image(), &jpeg.Options{Quality: 70})
			data = buffer.Bytes()
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
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
	engine.OnFullMatch("今日老公", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if rand.Float64() < 0.4 {
				uid := ctx.Event.UserID
				var husband int64 = zero.BotConfig.SuperUsers[rand.IntN(len(zero.BotConfig.SuperUsers))]
				avatarData, err := cache.GetAvatar(husband)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
				}
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你的群老公是"),
					message.ImageBytes(avatarData),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(husband), "]",
						"(", husband, ")哒",
					),
				)
			}
		})
	engine.OnFullMatch("今日猪头", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			botId := ctx.Event.SelfID
			pigsInfos, _ := qqwife.GetPigs(gid)
			pigs := []int64{}
			if len(pigsInfos) == 0 {
				// 获取群员列表
				qqgrouplist, err := cache.GetGroupMemberList(botId, ctx.Event.GroupID)
				if err != nil {
					ctx.SendError(err)
					return
				}

				qqgrouplist = qqgrouplist[len(qqgrouplist)/3:]
				pig_num := len(qqgrouplist)/20 + 1

				rand.Shuffle(len(qqgrouplist), func(i, j int) {
					qqgrouplist[i], qqgrouplist[j] = qqgrouplist[j], qqgrouplist[i]
				})
				pigs = qqgrouplist[:pig_num]
				err = qqwife.SavePigs(gid, pigs)
				if err != nil {
					ctx.SendError(err)
					return
				}
			} else {
				for _, v := range pigsInfos {
					pigs = append(pigs, v.UID)
				}
			}
			pigNum := rand.IntN(len(pigs))
			avatarData, err := cache.GetAvatar(pigs[pigNum])
			if err != nil {
				ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
				return
			}
			ctx.SendChain(
				message.At(uid),
				message.Text("今日的"+strconv.Itoa(pigNum)+"号猪头群友是"),
				message.ImageBytes(avatarData),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(pigs[pigNum]), "]",
					"(", pigs[pigNum], ")哒",
				),
			)
		})

	// 单身技能
	engine.OnRegex(`^(娶|嫁)\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			choice := ctx.State["regex_matched"].([]string)[1]
			fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("额,你的target好像不存在?"))
				return
			}
			selfId := ctx.Event.SelfID
			// 判断是否机器人自身
			if selfId == fiancee {
				return
			}
			res, msg := checkSingleDog(gid, uid, fiancee)
			if !res {
				if msg != "" {
					ctx.SendChain(message.Text(msg))
				}
				return
			}

			qqwife.Lock()
			defer qqwife.Unlock()
			// 写入CD
			err = qqwife.SaveCD(gid, uid, "嫁娶")
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			if uid == fiancee { // 如果是自己
				switch rand.IntN(3) {
				case 1:
					err := qqwife.SaveMarriageInfo(gid, uid, 0, "", "")
					if err != nil {
						ctx.SendChain(message.Text("[ERROR]:", err))
						return
					}
					ctx.SendChain(message.Text("今日获得成就：单身贵族"))
				default:
					ctx.SendChain(message.Text("今日获得成就：自恋狂"))
				}
				return
			}
			favor, err := qqwife.GetFavorability(uid, fiancee)
			if err != nil {
				ctx.SendError(err)
				return
			}
			if favor < 30 {
				favor = 30 // 保底30%概率
			}
			if rand.IntN(101) >= favor {
				ctx.SendChain(message.Text(sendtext[1][rand.IntN(len(sendtext[1]))]))
				return
			}

			favor, err = qqwife.UpdateFavorability(uid, fiancee, 5)
			if err != nil {
				ctx.SendError(err)
				return
			}

			// 去qqwife登记
			var choicetext string
			switch choice {
			case "娶":
				err := qqwife.SaveMarriageInfo(gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(fiancee))
				if err != nil {
					ctx.SendError(err)
					return
				}
				choicetext = "\n今天你的群老婆是"
			default:
				err := qqwife.SaveMarriageInfo(gid, fiancee, uid, ctx.CardOrNickName(fiancee), ctx.CardOrNickName(uid))
				if err != nil {
					ctx.SendError(err)
					return
				}
				choicetext = "\n今天你的群老公是"
			}

			go func() {
				// 请大家吃席
				avatarData, err := cache.GetAvatar(fiancee)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
					return
				}
				ctx.SendChain(
					message.Text(sendtext[0][rand.IntN(len(sendtext[0]))]),
					message.At(uid),
					message.Text(choicetext),
					message.ImageBytes(avatarData),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(fiancee), "]",
						"(", fiancee, ")哒\n当前你们好感度为", favor,
					),
				)
			}()

		})
	// NTR技能
	engine.OnRegex(`(^当(\[CQ:at,qq=(\d+)\]\s?|(\d+))的小三)|(^(ntr|牛头人|NTR)(\[CQ:at,qq=(\d+)\]\s?|(\d+)))`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var targetInfo UserInfo
			var ok bool
			if ok, targetInfo = checkMistress(ctx); !ok {
				return
			}

			qqwife.Lock()
			defer qqwife.Unlock()
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			fid := ctx.State["regex_matched"].([]string)
			target, e := strconv.ParseInt(fid[3]+fid[4], 10, 64)
			if e != nil {
				target, _ = strconv.ParseInt(fid[8]+fid[9], 10, 64)
			}
			// 写入CD
			err := qqwife.SaveCD(gid, uid, "NTR")
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			if target == uid {
				ctx.SendChain(message.Text("今日获得成就：自我攻略"))
				return
			}
			favor, err := qqwife.GetFavorability(uid, target)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			if favor < 30 {
				favor = 30 // 保底10%概率
			}
			// 最高33%
			if rand.IntN(101) >= favor/3 {
				ctx.SendChain(message.Text("失败了！可惜"))
				return
			}

			var choicetext string
			var greenID int64 // 被牛的

			err = qqwife.DelMarriageInfo(gid, target)
			if err != nil {
				ctx.SendChain(message.Text("ta不想和原来的对象分手...\n[error]", err))
				return
			}
			greenID = targetInfo.Target

			// 判断target是老公还是老婆
			if targetInfo.Mode == 1 {
				choicetext = "老公"
				err = qqwife.SaveMarriageInfo(gid, target, uid, ctx.CardOrNickName(target), ctx.CardOrNickName(uid))
			} else {
				choicetext = "老婆"
				err = qqwife.SaveMarriageInfo(gid, uid, target, ctx.CardOrNickName(uid), ctx.CardOrNickName(target))
			}
			if err != nil {
				ctx.SendError(err)
				return
			}

			_, err = qqwife.UpdateFavorability(uid, greenID, -5)
			if err != nil {
				ctx.SendError(err)
			}
			favor, err = qqwife.UpdateFavorability(uid, target, 5)
			if err != nil {
				ctx.SendError(err)
			}
			// 主要瓶颈就在网络i/o上
			go func() {
				avatarData, err := cache.GetAvatar(target)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
					return
				}
				ctx.SendChain(
					message.Text(sendtext[2][rand.IntN(len(sendtext[2]))]),
					message.At(uid),
					message.Text("今天你的群"+choicetext+"是"),
					message.ImageBytes(avatarData),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(target), "]",
						"(", target, ")哒\n当前你们好感度为", favor,
					),
				)
			}()
		})
	// 做媒技能
	engine.OnRegex(`^做媒\s?\[CQ:at,qq=(\d+)\]\s?\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup, checkMatchmaker).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			gayOne, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			gayZero, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			qqwife.Lock()
			defer qqwife.Unlock()
			// 写入CD
			err := qqwife.SaveCD(gid, uid, "做媒")
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			favor, err := qqwife.GetFavorability(gayOne, gayZero)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			if favor < 30 {
				favor = 30 // 保底30%概率
			}
			if rand.IntN(101) >= favor {
				ctx.SendChain(message.Text(sendtext[1][rand.IntN(len(sendtext[1]))]))
				return
			}
			// 去qqwife登记
			err = qqwife.SaveMarriageInfo(gid, gayOne, gayZero, ctx.CardOrNickName(gayOne), ctx.CardOrNickName(gayZero))
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			_, err = qqwife.UpdateFavorability(uid, gayOne, 1)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
			}
			_, err = qqwife.UpdateFavorability(uid, gayZero, 1)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
			}
			_, err = qqwife.UpdateFavorability(gayOne, gayZero, 1)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
			}
			// 请大家吃席
			go func() {
				avatarData, err := cache.GetAvatar(gayZero)
				if err != nil {
					ctx.SendChain(message.Text(fmt.Sprint("[redis]", err)))
					return
				}
				ctx.SendChain(
					message.At(uid),
					message.Text("恭喜你成功撮合了一对CP\n\n"),
					message.At(gayOne),
					message.Text("今天你的群老婆是"),
					message.ImageBytes(avatarData),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(gayZero), "]",
						"(", gayZero, ")哒",
					),
				)
			}()
		})
	engine.OnFullMatchGroup([]string{"闹离婚", "办离婚"}, zero.OnlyGroup, checkDivorce).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID

			// 判断是否符合条件
			userInfo, _ := qqwife.GetMarriageInfo(gid, uid)
			if userInfo == (UserInfo{}) { // 如果空数据
				ctx.SendChain(message.Text("今天你还没结婚哦"))
				return
			}

			// 写入CD
			err := qqwife.SaveCD(gid, uid, "离婚")
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}

			target := userInfo.Target

			favor, err := qqwife.GetFavorability(uid, target)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			if favor < 20 {
				favor = 10
			}
			// 离婚失败
			if rand.IntN(101) > 110-favor {
				ctx.SendChain(message.Text(sendtext[3][rand.IntN(len(sendtext[3]))]))
				return
			}
			err = qqwife.DelMarriageInfo(gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Text(sendtext[4][rand.IntN(len(sendtext[4]))]))
		})

}

func slicename(name string, canvas *gg.Context) (resultname string) {
	usermane := []rune(name) // 将每个字符单独放置
	widthlen := 0
	numberlen := 0
	for i, v := range usermane {
		width, _ := canvas.MeasureString(string(v)) // 获取单个字符的宽度
		widthlen += int(width)
		if widthlen > 350 {
			break // 总宽度不能超过350
		}
		numberlen = i
	}
	if widthlen > 350 {
		resultname = string(usermane[:numberlen-1]) + "......" // 名字切片
	} else {
		resultname = name
	}
	return
}
