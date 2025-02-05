// 魔改自 https://github.com/azmiao/pcrjjc3-tw
package pcrjjc3

import (
	zero "ZeroBot"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ZeroBot/extension"
	"ZeroBot/message"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

var (
	pluginName = "pcrjjc"
	help       = `注意：数字2为服务器编号，仅支持2~4服
[竞技场bind 10位uid] 默认双场均启用，排名下降时推送
[竞技场查询 10位uid] 查询（bind后无需输入uid，可缩写为jjccx、看看）
[停止竞技场bind] 停止jjc推送
[停止公主竞技场bind] 停止pjjc推送
[启用竞技场bind] 启用jjc推送
[启用公主竞技场bind] 启用pjjc推送
[竞技场历史] jjc变化记录（bind开启有效，可保留10条）
[公主竞技场历史] pjjc变化记录（bind开启有效，可保留10条）
[详细查询 10位uid] 能不用就不用（bind后无需输入2 uid） 
[竞技场关注 10位uid] 默认双场均启用，排名变化及上线时推送 
[删除bind] 删除bind
[删除关注 x] 删除第x个关注
[竞技场bind状态] 查看排名变动推送bind状态
[关注列表] 返回关注的序号以及对应的游戏UID
[关注查询 x] 查询第x个关注 可缩写为看看`
	userInfoManage = UserInfoManage{update: true}
	config         = Config{}
)

func init() {
	// Read config YAML file
	data, err := os.ReadFile(filepath.Join(dataPath, "config.yaml"))
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return
	}

	// Unmarshal YAML data
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
	}
	proxy = config.Proxy

	patternId := `\s*(\d{10})\s*`
	reId := regexp.MustCompile(patternId)
	engine := zero.NewTemplate(&zero.Metadata{
		Name: pluginName,
		Help: help,
	})
	engine.OnPrefixGroup([]string{"看看", "jjccx", "关注查询"}).Handle(
		func(ctx *zero.Ctx) {
			id := ""
			uid := fmt.Sprint(ctx.Event.UserID)
			arg := ctx.State["args"].(string)

			// 查找匹配
			match := reId.FindStringSubmatch(arg)
			if len(match) == 2 {
				id = match[1]
			}

			var msg string
			if id == "" {
				for _, segment := range ctx.Event.Message {
					if segment.Type == "at" {
						uid = segment.Data["qq"]
					}
				}
				order := 0
				orderStr := strings.TrimSpace(arg)
				if orderStr != "" {
					order, err = strconv.Atoi(orderStr)
					if err != nil {
						return
					}
					if order > 100 {
						return
					}
				}
				userMap := userInfoManage.UserInfoMap
				if userInfo, ok := userMap[uid]; ok {
					if len(userInfo.Id) <= order {
						msg = "关注序号错误，超出上限"
					} else if order == 0 && userInfo.Id[0] == "" {
						msg = "该用户没有绑定账号"
					} else {
						id = userInfo.Id[order]
					}
				} else {
					msg = "该用户没有绑定账号"
				}
			}

			ctx.Stop()
			if id != "" {
				msg, err = userQuery(id)
				if err != nil {
					ctx.SendError(err)
					return
				}
			}

			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
		})

	engine.OnRegex(`竞技场bind\s*(\d)\s*(\d{9})$`).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			model := extension.RegexModel{}
			_ = ctx.Parse(&model)
			match := model.Matched
			cx := match[1]
			oldId := match[2]
			uid := strconv.FormatInt(ctx.Event.UserID, 10)
			bid := strconv.FormatInt(ctx.Event.SelfID, 10)
			gid := strconv.FormatInt(ctx.Event.GroupID, 10)
			if gid == "0" {
				gid = ""
			}
			msg, err := userInfoManage.bind(cx+oldId, uid, gid, bid, false)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
				msg = err.Error()
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
		})

	engine.OnRegex(`竞技场关注\s*(\d)\s*(\d{9})$`).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			model := extension.RegexModel{}
			_ = ctx.Parse(&model)
			match := model.Matched
			cx := match[1]
			oldId := match[2]
			uid := strconv.FormatInt(ctx.Event.UserID, 10)
			bid := strconv.FormatInt(ctx.Event.SelfID, 10)
			gid := strconv.FormatInt(ctx.Event.GroupID, 10)
			if gid == "0" {
				gid = ""
			}
			msg, err := userInfoManage.bind(cx+oldId, uid, gid, bid, true)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
				msg = err.Error()
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
		})

	engine.OnFullMatch("关注列表").SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			uid := strconv.FormatInt(ctx.Event.UserID, 10)
			msg := userInfoManage.attentionList(uid)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
		})

	engine.OnPrefixGroup([]string{"删除绑定", "删除bind"}).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			atInfos := ctx.GetAtInfos()
			if zero.SuperUserPermission(ctx) && len(atInfos) > 0 {
				for _, atInfo := range atInfos {
					uid := strconv.FormatInt(atInfo.QQ, 10)
					msg := userInfoManage.delBind(uid, 0)
					ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
				}
				return
			}
			uid := strconv.FormatInt(ctx.Event.UserID, 10)
			msg := userInfoManage.delBind(uid, 0)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
		})

	engine.OnRegex(`删除关注\s*(\d+)`).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			model := extension.RegexModel{}
			_ = ctx.Parse(&model)
			match := model.Matched
			num, _ := strconv.Atoi(match[1])
			uid := strconv.FormatInt(ctx.Event.UserID, 10)
			msg := userInfoManage.delBind(uid, num)
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
		})

	engine.OnPrefix("更新版本", zero.SuperUserPermission).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			model := extension.PrefixModel{}
			_ = ctx.Parse(&model)
			version := model.Args
			err := updateVersion(version)
			msg := ""
			if err != nil {
				msg = "[pcr] " + err.Error()
			} else {
				msg = "更新版本成功"
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
		})

	engine.OnRegex(`(开启|关闭)轮询`, zero.SuperUserPermission).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			model := extension.RegexModel{}
			_ = ctx.Parse(&model)
			match := model.Matched
			if match[1] == "开启" {
				config.setGloablPush(true)
				ctx.Send("已开启轮询")
			} else {
				config.setGloablPush(false)
				ctx.Send("已关闭轮询")
			}
		})

	// 定时任务
	go func() {
		for range time.NewTicker(time.Duration(config.ScheduleTime) * time.Second).C {

			if !config.GlobalPush {
				continue
			}

			startTime := time.Now()
			list := userInfoManage.GetIdList()

			queryAll(list, config.ScheduleThread)
			sendChange()

			log.Info().Str("name", pluginName).Dur("轮询时间", time.Since(startTime)).Msg("")
		}
	}()
}

func sendChange() {
	userInfoManage.RLock()
	for uid, userInfo := range userInfoManage.UserInfoMap {
		for i, id := range userInfo.Id {
			if id == "" {
				continue
			}
			msg := getChange(id, i, userInfo.Mode[i])
			if msg != "" {
				bid, _ := strconv.ParseInt(userInfo.Bid[i], 10, 64)
				bot, err := zero.GetBot(bid)
				if err != nil {
					log.Error().Str("name", pluginName).Err(err).Msg("")
					continue
				}
				qq, _ := strconv.ParseInt(uid, 10, 64)
				if userInfo.Gid[i] == "" {
					bot.SendPrivateMessage(qq, message.Message{message.At(qq), message.Text(msg)})
				} else {
					gid, _ := strconv.ParseInt(userInfo.Gid[i], 10, 64)
					bot.SendGroupMessage(gid, message.Message{message.At(qq), message.Text(msg)})
				}
			}
		}
	}
	userInfoManage.RUnlock()

	gamerInfoManage.Lock()
	gamerInfoManage.GamerInfoMap = gamerInfoManage.TmpGamerInfoMap
	gamerInfoManage.Unlock()
}
