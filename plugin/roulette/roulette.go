package roulette

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	zero "ZeroBot"
	"ZeroBot/message"
)

type GameMode int

const (
	ModeKick GameMode = 0
	ModeBan  GameMode = 1
)

type GameStatus struct {
	Mode         GameMode
	BulletPos    int // 1-6
	CurrentCount int
	LastActive   time.Time
}

var (
	gameMap    = make(map[int64]*GameStatus)
	banPlayers = make(map[int64][]int64) // groupID -> []userID
	mu         sync.Mutex
)

var (
	metaData = &zero.Metadata{
		Name: "轮盘赌",
		Help: "- 牛牛轮盘 (默认禁言模式)\n- 牛牛轮盘踢人\n- 牛牛轮盘禁言\n- 牛牛开枪\n- 牛牛救一下 [@用户]\n- 牛牛补一枪 [@用户]",
	}
	engine  = zero.NewTemplate(metaData)
	timeout = 300 * time.Second
)

var shotText = []string{
	"无需退路。",
	"英雄们啊，为这最强大的信念，请站在我们这边。",
	"颤抖吧，在真正的勇敢面前。",
	"哭嚎吧，为你们不堪一击的信念。",
	"现在可没有后悔的余地了。",
	"你将在此跪拜。",
}

func init() {
	engine.OnRegex(`^牛牛轮盘\s?(踢人|禁言)?$`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()

		gid := ctx.Event.GroupID
		if gid == 0 {
			ctx.SendChain(message.Text("请在群聊中使用此命令"))
			return
		}

		status, ok := gameMap[gid]
		if ok && time.Since(status.LastActive) < timeout {
			ctx.SendChain(message.Text("游戏已经在进行中，或者刚刚结束，请稍后再试。"))
			return
		}

		modeStr := ctx.State["regex_matched"].([]string)[1]
		mode := ModeBan
		if modeStr == "踢人" {
			mode = ModeKick
		}

		gameMap[gid] = &GameStatus{
			Mode:         mode,
			BulletPos:    rand.IntN(6) + 1,
			CurrentCount: 0,
			LastActive:   time.Now(),
		}

		// 检查 bot 权限
		info, err := ctx.GetThisGroupMemberInfo(ctx.Event.SelfID, false)
		if err == nil {
			role := info.Get("role").String()
			if role != "admin" && role != "owner" {
				ctx.SendChain(message.Text("虽然我很想陪你们玩，但我现在还没有管理员权限呢..."))
				delete(gameMap, gid)
				return
			}
		}

		typeMsg := "禁言"
		if mode == ModeKick {
			typeMsg = "踢出群聊"
		}

		ctx.SendChain(message.Text(fmt.Sprintf("这是一把充满荣耀与死亡的左轮手枪，六个弹槽只有一颗子弹，中弹的那个人将会被%s。勇敢的战士们啊，扣动你们的扳机吧！", typeMsg)))
	})

	engine.OnFullMatch("牛牛开枪", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()

		gid := ctx.Event.GroupID
		status, ok := gameMap[gid]
		if !ok || time.Since(status.LastActive) > timeout {
			ctx.SendChain(message.Text("当前没有正在进行的游戏，请发送“牛牛轮盘”开始。"))
			return
		}

		status.CurrentCount++
		status.LastActive = time.Now()

		if status.CurrentCount == status.BulletPos {
			// 中弹
			delete(gameMap, gid)
			uid := ctx.Event.UserID

			if status.Mode == ModeKick {
				ctx.SendChain(message.Text("米诺斯英雄们的故事......有喜剧，便也会有悲剧。舍弃了荣耀，"), message.At(uid), message.Text("选择回归平凡......"))
				// 踢人
				ctx.SetThisGroupKick(uid, false)
			} else {
				duration := int64(rand.IntN(15)+5) * 60
				ctx.SendChain(message.Text("米诺斯英雄们的故事......有喜剧，便也会有悲剧。舍弃了荣耀，"), message.At(uid), message.Text(fmt.Sprintf("选择回归平凡...... (禁言 %d 分钟)", duration/60)))
				// 禁言
				ctx.SetThisGroupBan(uid, duration)
				banPlayers[gid] = append(banPlayers[gid], uid)
			}
		} else {
			// 未中弹
			if status.CurrentCount >= 6 {
				// 炸膛概率 (参考 Python 代码 0.125)
				if rand.Float32() < 0.125 {
					delete(gameMap, gid)
					ctx.SendChain(message.Text("我的手中的这把武器，找了无数工匠都难以修缮如新。不......不该如此......"))
					return
				}
			}
			msg := shotText[status.CurrentCount-1]
			ctx.SendChain(message.Text(fmt.Sprintf("%s ( %d / 6 )", msg, status.CurrentCount)))

			// Bot 概率参与 (参考 Python 代码 0.1667)
			if rand.Float32() < 0.1667 {
				time.Sleep(1 * time.Second)
				ctx.SendChain(message.Text("我也来试试运气..."))
				status.CurrentCount++
				status.LastActive = time.Now()
				if status.CurrentCount == status.BulletPos {
					delete(gameMap, gid)
					ctx.SendChain(message.Text("砰！看来我的运气不太好呢..."))
					if status.Mode == ModeKick {
						// Bot 没法踢自己，发个表情或者撤退
						ctx.SendChain(message.Text("米诺斯英雄们的故事......有喜剧，便也会有悲剧。舍弃了荣耀，我选择回归平凡...... (Bot 已退出群聊)"))
						ctx.SetThisGroupLeave(false)
					} else {
						ctx.SendChain(message.Text("虽然我没法禁言自己，但我会安静一会儿的... (Bot 进入静默模式)"))
					}
				} else {
					msg := shotText[status.CurrentCount-1]
					ctx.SendChain(message.Text(fmt.Sprintf("呼... 看来我还命不该绝。%s ( %d / 6 )", msg, status.CurrentCount)))
				}
			}
		}
	})

	engine.OnPrefix("牛牛救一下", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			return
		}

		atInfos := ctx.GetAtInfos()
		if len(atInfos) > 0 {
			// 救指定的人
			for _, at := range atInfos {
				ctx.SetThisGroupBan(at.QQ, 0)
			}
			ctx.SendChain(message.Text("命运之手指向了为沉默所困之人，已从沉默中被解放。"))
		} else {
			// 救所有人 (在本插件记录中的)
			mu.Lock()
			players := banPlayers[gid]
			delete(banPlayers, gid)
			mu.Unlock()

			if len(players) == 0 {
				ctx.SendChain(message.Text("此刻并无需要拯救之人，和平仍在延续。"))
				return
			}

			for _, uid := range players {
				ctx.SetThisGroupBan(uid, 0)
			}
			ctx.SendChain(message.Text("命运的轮盘再次转动，所有的沉默都被打破。"))
		}
	})

	engine.OnPrefix("牛牛补一枪", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			return
		}

		atInfos := ctx.GetAtInfos()
		if len(atInfos) > 0 {
			// 补指定的人
			for _, at := range atInfos {
				duration := int64(rand.IntN(30)+30) * 60
				ctx.SetThisGroupBan(at.QQ, duration)
			}
			ctx.SendChain(message.Text("哭嚎吧，为你们不堪一击的信念。"))
		} else {
			// 补所有人 (在本插件记录中的)
			mu.Lock()
			players := banPlayers[gid]
			mu.Unlock()

			if len(players) == 0 {
				ctx.SendChain(message.Text("转身吧，勇士们。我们已经获得了完美的胜利，现在是该回去享受庆祝的盛典了。"))
				return
			}

			for _, uid := range players {
				duration := int64(rand.IntN(20)+20) * 60
				ctx.SetThisGroupBan(uid, duration)
			}
			ctx.SendChain(message.Text("是吗，我们做到了吗......我现在，正体会至高的荣誉和幸福。"))
		}
	})
}
