// Package control 控制插件的启用与优先级等
package manager

import (
	"MiniBot/plugin/manager/plugin"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	zero "ZeroBot"
	"ZeroBot/extension"
	"ZeroBot/message"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

const (
	// StorageFolder 插件控制数据目录
	StorageFolder = "data/control/"
	// Md5File ...
	Md5File = StorageFolder + "stor.spb"
	dbfile  = StorageFolder + "plugins.db"
	lnfile  = StorageFolder + "lnperpg.txt"
)

type Limiter struct {
	rl          *rate.Limiter
	expiredTime time.Time
}

type LimiterManger struct {
	limiterMap map[int64]*Limiter
	sync.RWMutex
}

func (LM *LimiterManger) GetLimiter(key int64) *Limiter {
	LM.RLock()
	defer LM.RUnlock()
	return LM.limiterMap[key]
}

func (LM *LimiterManger) NewLimiter(key int64) *Limiter {
	limiter := &Limiter{
		rl:          rate.NewLimiter(3, 5),
		expiredTime: time.Now().Add(24 * time.Hour),
	}
	LM.Lock()
	defer LM.Unlock()
	LM.limiterMap[key] = limiter
	return limiter
}

func (LM *LimiterManger) GetOrNewLimiter(key int64) *Limiter {
	limter := LM.GetLimiter(key)
	if limter == nil {
		limter = LM.NewLimiter(key)
	}
	return limter
}

func (LM *LimiterManger) CleanLimiter() {
	LM.Lock()
	defer LM.Unlock()
	for key, limter := range LM.limiterMap {
		if time.Now().After(limter.expiredTime) {
			delete(LM.limiterMap, key)
		}
	}
}

var (
	limterCount = 0
	managers    = NewManager()
	LM          = &LimiterManger{
		limiterMap: map[int64]*Limiter{},
	}
)

// engine 级
func newAuthHandler(metaDate *zero.Metadata) zero.Rule {
	c := plugin.CM.NewControl(metaDate)
	return func(ctx *zero.Ctx) bool {
		// 对超管无用
		if zero.SuperUserPermission(ctx) {
			return true
		}

		bid := ctx.Event.SelfID
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		uidGobal := strconv.FormatInt(uid, 10)
		uidKey := bidWithuid(bid, uid)
		gidKey := bidWithgid(bid, gid)
		bidKey := strconv.FormatInt(bid, 10)

		// 对群，全局个人，bot级个人，机器人，所有机器人的权限情况 可以理解为这个为全插件级,排除默认插件
		if managers.IsBlocked(gidKey) || managers.IsBlocked(uidGobal) || managers.IsBlocked(uidKey) || managers.IsBlocked(bidKey) || managers.IsBlocked("0") {
			ctx.Stop()
			return false
		}

		controlLevel := c.MetaDate.Level
		if controlLevel > 0 {
			// 判断权限等级是否足够
			levels := []uint{}
			levels = append(levels, managers.GetLevel(uidKey))
			if gid != 0 {
				levels = append(levels, managers.GetLevel(gidKey))
			}
			// 取最大值
			level := slices.Max(levels)

			if level < controlLevel {
				return false
			}
		}

		// 该插件针对所有机器人， 机器人，群，个人 的权限情况
		return c.IsEnabled("0") && c.IsEnabled(bidKey) && c.IsEnabled(gidKey) && c.IsEnabled(uidKey)
	}
}

func judgeMalicious(ctx *zero.Ctx) {
	uid := ctx.Event.UserID
	uidGobal := strconv.FormatInt(uid, 10)
	// 恶意用户，封禁级别为全局
	if MC.BlockMalicious {
		limiter := LM.GetOrNewLimiter(uid)
		if !limiter.rl.Allow() {
			managers.DoBlock(uidGobal)
			log.Info().Str("name", pluginName).Msgf("[manager] 封禁恶意用户uid%d", uid)
		}

		// 每10000次触发，删除
		limterCount++
		if limterCount == 10000 {
			LM.CleanLimiter()
			limterCount = 0
		}
	}
}

func init() {
	zero.OnCommandGroup([]string{
		"响应", "response", "沉默", "silence",
	}, zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		bid := ctx.Event.SelfID
		gidKey := bidWithgid(bid, gid)
		uidKey := bidWithuid(bid, uid)
		var msg string
		switch ctx.State["command"] {
		case "响应", "response":
			if gid == 0 {
				managers.DoUnblock(uidKey)
			} else {
				managers.DoUnblock(gidKey)
			}
			msg = zero.BotConfig.NickName[0] + "将开始在此工作啦~"
		case "沉默", "silence":
			if gid == 0 {
				managers.DoBlock(uidKey)
			} else {
				managers.DoBlock(gidKey)
			}
			msg = zero.BotConfig.NickName[0] + "将开始休息啦~"
		default:
			msg = "ERROR: bad command"
		}
		ctx.SendChain(message.Text(msg))
	})

	zero.OnCommandGroup([]string{
		"全局响应", "allresponse", "全局沉默", "allsilence",
	}, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		bid := ctx.Event.SelfID
		bidKey := strconv.FormatInt(bid, 10)
		var msg message.MessageSegment
		cmd := ctx.State["command"].(string)
		switch {

		case strings.Contains(cmd, "响应") || strings.Contains(cmd, "response"):
			managers.DoUnblock(bidKey)
			msg = message.Text(zero.BotConfig.NickName[0], "将开始在全部位置工作啦~")
		case strings.Contains(cmd, "沉默") || strings.Contains(cmd, "silence"):
			managers.DoBlock(bidKey)
			msg = message.Text(zero.BotConfig.NickName[0], "将开始休息啦~")
		default:
			msg = message.Text("ERROR: bad command\"", cmd, "\"")
		}
		ctx.SendChain(msg)
	})

	zero.OnCommandGroup([]string{
		"启用", "enable", "禁用", "disable",
	}, zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		service, ok := plugin.CM.Lookup(model.Args)
		if !ok {
			ctx.SendChain(message.Text("没有找到指定服务!"))
			return
		}
		gid := ctx.Event.GroupID
		bid := ctx.Event.SelfID
		uid := ctx.Event.UserID

		gidKey := bidWithgid(bid, gid)
		uidKey := bidWithuid(bid, uid)

		if strings.Contains(model.Command, "启用") || strings.Contains(model.Command, "enable") {
			if gid == 0 {
				service.Enable(uidKey)
			} else {
				service.Enable(gidKey)
			}
			ctx.SendChain(message.Text("已启用服务: " + model.Args))
		} else {
			if gid == 0 {
				service.Disable(uidKey)
			} else {
				service.Disable(gidKey)
			}
			ctx.SendChain(message.Text("已禁用服务: " + model.Args))
		}
	})

	zero.OnRegex(`^设置权限等级\s*(\d+)\s*(\d*)`, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.RegexModel{}
		_ = ctx.Parse(&model)
		bid := ctx.Event.SelfID
		level, _ := strconv.ParseInt(model.Matched[1], 10, 64)
		uid, err := strconv.ParseInt(model.Matched[2], 10, 64)
		msg := ""
		if err == nil {
			uidKey := bidWithuid(bid, uid)
			managers.SetLevel(uidKey, uint(level))
			msg = fmt.Sprintf("设置用户%d权限等级为%d", uid, level)
		} else {
			gid := ctx.Event.GroupID
			gidKey := bidWithgid(bid, gid)
			managers.SetLevel(gidKey, uint(level))
			msg = fmt.Sprintf("设置群%d权限等级为%d", gid, level)
		}

		ctx.SendChain(message.Text(msg))

	})

	zero.OnCommandGroup([]string{
		"全局启用", "allenable", "全局禁用", "alldisable",
	}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		service, ok := plugin.CM.Lookup(model.Args)
		if !ok {
			ctx.SendChain(message.Text("没有找到指定服务!"))
			return
		}
		bid := ctx.Event.SelfID
		bidKey := strconv.FormatInt(bid, 10)
		if strings.Contains(model.Command, "启用") || strings.Contains(model.Command, "enable") {
			service.Enable(bidKey)
			ctx.SendChain(message.Text("已全局启用服务: " + model.Args))
		} else {
			service.Disable(bidKey)
			ctx.SendChain(message.Text("已全局禁用服务: " + model.Args))
		}
	})

	// zero.OnCommandGroup([]string{"还原", "reset"}, zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
	// 	model := extension.CommandModel{}
	// 	_ = ctx.Parse(&model)
	// 	service, ok := Lookup(model.Args)
	// 	if !ok {
	// 		ctx.SendChain(message.Text("没有找到指定服务!"))
	// 		return
	// 	}
	// 	gid := ctx.Event.GroupID
	// 	if gid == 0 {
	// 		// 个人用户
	// 		gid = -ctx.Event.UserID
	// 	}
	// 	service.Reset(gid)
	// 	ctx.SendChain(message.Text("已还原服务的默认启用状态: " + model.Args))
	// })

	zero.OnCommandGroup([]string{
		"封禁", "block", "解封", "unblock",
	}, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		args := strings.Split(model.Args, " ")
		bid := ctx.Event.SelfID
		if len(args) >= 1 {
			msg := "**报告**"
			if strings.Contains(model.Command, "解") || strings.Contains(model.Command, "un") {
				for _, usr := range args {
					uid, err := strconv.ParseInt(usr, 10, 64)
					uidKey := bidWithuid(bid, uid)
					if err == nil {
						managers.DoUnblock(uidKey)
					}
				}
			} else {
				for _, usr := range args {
					uid, err := strconv.ParseInt(usr, 10, 64)
					uidKey := bidWithuid(bid, uid)
					if err == nil {
						managers.DoBlock(uidKey)
					}
				}
			}
			ctx.SendChain(message.Text(msg))
			return
		}
		ctx.SendChain(message.Text("参数错误!"))
	})

	if MC.BlockStranger {
		// 使用中间件级
		zero.GolbaleMiddleware.Use(
			func(ctx *zero.Ctx) {
				if ctx.Event.MessageType == "private" && ctx.Event.SubType != "friend" {
					ctx.Stop()
					log.Debug().Str("name", pluginName).Msg("阻止了非好友的私聊")
				}
			},
		)
	}
}
