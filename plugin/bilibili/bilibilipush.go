// Package bilibili b站推送
package bilibili

import (
	"MiniBot/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	zero "ZeroBot"
	"ZeroBot/message"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const (
	ua      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	referer = "https://space.bilibili.com/%v"
	infoURL = "https://api.bilibili.com/x/space/wbi/acc/info?mid=%v"
)

// bdb bilibili推送数据库
var bdb *bilibilipushdb

var (
	lastTime   = map[int64]int64{}
	liveStatus = map[int64]int{}
	upMap      = map[int64]string{}
)

func init() {
	en := zero.NewTemplate(&zero.Metadata{
		Name: "b站推送",
		Help: "- 添加b站订阅[uid|name]\n" +
			"- 取消b站订阅[uid|name]\n" +
			"- 取消b站动态订阅[uid|name]\n" +
			"- 取消b站直播订阅[uid|name]\n" +
			"- b站推送列表\n" +
			"- [开启|关闭]艾特全体\n" +
			"Tips: 需要先在 bilibili 插件中设置cookie\n",
		Level: 2,
	})

	// 加载bilibili推送数据库
	bdb = initializePush()
	en.OnFullMatch(`开启艾特全体`, zero.UserOrGrpAdmin, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if err := changeAtAll(gid, 1); err != nil {
			ctx.SendError(err)
			return
		}
		ctx.SendChain(message.Text("已开启艾特全体Oo"))
	})

	en.OnFullMatch(`关闭艾特全体`, zero.UserOrGrpAdmin, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if err := changeAtAll(gid, 0); err != nil {
			ctx.SendError(err)
			return
		}
		ctx.SendChain(message.Text("已关闭艾特全体Oo"))
	})

	en.OnRegex(`^添加[B|b]站订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		name, err := getName(buid, cfg)
		if err != nil || name == "" {
			ctx.SendError(err)
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := subscribe(buid, gid, ctx.Event.SelfID); err != nil {
			ctx.SendError(err)
			return
		}
		ctx.SendChain(message.Text("已添加" + name + "的订阅"))
	})

	en.OnRegex(`^取消[B|b]站订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		name, err := getName(buid, cfg)
		if err != nil {
			ctx.SendError(err)
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := unsubscribe(buid, gid, ctx.Event.SelfID); err != nil {
			ctx.SendError(err)
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的订阅"))
	})
	en.OnRegex(`^取消[B|b]站动态订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		name, err := getName(buid, cfg)
		if err != nil {
			ctx.SendError(err)
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := unsubscribeDynamic(buid, gid, ctx.Event.SelfID); err != nil {
			ctx.SendError(err)
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的动态订阅"))
	})
	en.OnRegex(`^取消[B|b]站直播订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		name, err := getName(buid, cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if err := unsubscribeLive(buid, gid, ctx.Event.SelfID); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的直播订阅"))
	})

	en.OnRegex(`^[B|b]站推送列表$`, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		bpl := bdb.getAllPushByGroup(gid)
		msg := "--------B站推送列表--------"
		for _, v := range bpl {
			if _, ok := upMap[v.BilibiliUID]; !ok {
				bdb.updateAllUp()
			}
			msg += fmt.Sprintf("\nuid:%-12d 动态：", v.BilibiliUID)
			if v.DynamicDisable == 0 {
				msg += "●"
			} else {
				msg += "○"
			}
			msg += " 直播："
			if v.LiveDisable == 0 {
				msg += "●"
			} else {
				msg += "○"
			}
			msg += " up主：" + upMap[v.BilibiliUID]
		}
		data, err := text.RenderToBase64(msg, text.FontFile, 600, 20)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.ImageBytes(data))
	})

	// 定时任务
	go func() {
		for range time.NewTicker(120 * time.Second).C {
			err := sendDynamic()
			if err != nil {
				log.Error().Str("name", "b站推送").Err(err).Msg("")
			}
			err = sendLive()
			if err != nil {
				log.Error().Str("name", "b站推送").Err(err).Msg("")
			}
		}
	}()
}

func changeAtAll(gid int64, b int) (err error) {
	return bdb.updateAtAll(bilibiliAt{
		GroupID: gid,
		AtAll:   b,
	})
}

// 取得uid的名字
func getName(buid int64, cookiecfg *bz.CookieConfig) (name string, err error) {
	var ok bool
	if name, ok = upMap[buid]; !ok {
		data, err := web.RequestDataWithHeaders(web.NewDefaultClient(), bz.SignURL(fmt.Sprintf(infoURL, buid)), "GET", func(r *http.Request) error {
			if cookiecfg != nil {
				cookie := ""
				cookie, err = cookiecfg.Load()
				if err != nil {
					return err
				}
				r.Header.Add("Cookie", cookie)
			}
			r.Header.Set("User-Agent", ua)
			return nil
		}, nil)
		if err != nil {
			return "", err
		}
		dataJson := gjson.ParseBytes(data)
		status := dataJson.Get("code").Int()
		if status != 0 {
			err = errors.New(dataJson.Get("message").String())
			return "", err
		}
		name = dataJson.Get("data.name").String()
		bdb.insertBilibiliUp(buid, name)
		upMap[buid] = name
	}
	return
}

// subscribe 订阅
func subscribe(buid, groupid, SelfID int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"live_disable":    0,
		"dynamic_disable": 0,
		"bot_id":          SelfID,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

// unsubscribe 取消订阅
func unsubscribe(buid, groupid int64, SelfID int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"live_disable":    1,
		"dynamic_disable": 1,
		"bot_id":          SelfID,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

func unsubscribeDynamic(buid, groupid int64, SelfID int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"dynamic_disable": 1,
		"bot_id":          SelfID,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

func unsubscribeLive(buid, groupid int64, SelfID int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid": buid,
		"group_id":     groupid,
		"live_disable": 1,
		"bot_id":       SelfID,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

func getUserDynamicCard(buid int64, cookiecfg *bz.CookieConfig) (cardList []gjson.Result, err error) {
	data, err := web.RequestDataWithHeaders(web.NewDefaultClient(), fmt.Sprintf(bz.SpaceHistoryURL, buid, 0), "GET", func(req *http.Request) error {
		if cookiecfg != nil {
			cookie := ""
			cookie, err = cookiecfg.Load()
			if err != nil {
				return err
			}
			req.Header.Add("Cookie", cookie)
		}
		req.Header.Add("User-Agent", ua)
		return nil
	}, nil)
	if err != nil {
		return
	}
	cardList = gjson.GetBytes(data, "data.cards").Array()
	return
}

func getLiveList(uids ...int64) (string, error) {
	m := make(map[string]any)
	m["uids"] = uids
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	data, err := web.PostData(bz.LiveListURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	return utils.BytesToString(data), nil
}

func sendDynamic() error {
	uids := bdb.getAllBuidByDynamic()
	for _, buid := range uids {
		time.Sleep(2 * time.Second)
		cardList, err := getUserDynamicCard(buid, cfg)
		if err != nil {
			return err
		}
		if len(cardList) == 0 {
			return nil
		}
		t, ok := lastTime[buid]
		// 第一次先记录时间,啥也不做
		if !ok {
			lastTime[buid] = cardList[0].Get("desc.timestamp").Int()
			return nil
		}
		for i := len(cardList) - 1; i >= 0; i-- {
			ct := cardList[i].Get("desc.timestamp").Int()
			if ct > t && ct > time.Now().Unix()-600 {
				lastTime[buid] = ct
				infoList := bdb.getInfoByBuidAndDynamic(buid)
				dc, err := bz.LoadDynamicDetail(cardList[i].Raw)
				if err != nil {
					err = errors.Errorf("动态%v的解析有问题,%v", cardList[i].Get("desc.dynamic_id_str"), err)
					return err
				}
				msg, err := dynamicCard2msg(&dc)
				if err != nil {
					err = errors.Errorf("动态%v的解析有问题,%v", cardList[i].Get("desc.dynamic_id_str"), err)
					return err
				}
				for _, info := range infoList {
					time.Sleep(time.Millisecond * 100)
					bot, err := zero.GetBot(info.BotID)
					if err != nil {
						log.Error().Str("name", "b站推送").Err(err).Msg("")
					}
					switch {
					case info.GroupID > 0:
						bot.SendGroupMessage(info.GroupID, msg)
					case info.GroupID < 0:
						bot.SendPrivateMessage(-info.GroupID, msg)
					}
				}
			}
		}
	}
	return nil
}

func sendLive() error {
	uids := bdb.getAllBuidByLive()
	ll, err := getLiveList(uids...)
	if err != nil {
		return err
	}
	gjson.Get(ll, "data").ForEach(func(key, value gjson.Result) bool {
		newStatus := int(value.Get("live_status").Int())
		if newStatus == 2 {
			newStatus = 0
		}
		if _, ok := liveStatus[key.Int()]; !ok {
			liveStatus[key.Int()] = newStatus
			return true
		}
		oldStatus := liveStatus[key.Int()]
		if newStatus != oldStatus && newStatus == 1 {
			liveStatus[key.Int()] = newStatus
			infoList := bdb.getInfoByBuidAndLive(key.Int())
			roomID := value.Get("short_id").Int()
			if roomID == 0 {
				roomID = value.Get("room_id").Int()
			}
			lURL := bz.LiveURL + strconv.FormatInt(roomID, 10)
			lName := value.Get("uname").String()
			lTitle := value.Get("title").String()
			lCover := value.Get("cover_from_user").String()
			if lCover == "" {
				lCover = value.Get("keyframe").String()
			}
			var msg []message.MessageSegment
			msg = append(msg, message.Text(lName+" 正在直播：\n"))
			msg = append(msg, message.Text(lTitle))
			msg = append(msg, message.Image(lCover))
			msg = append(msg, message.Text("直播链接：", lURL))

			for _, info := range infoList {
				bot, err := zero.GetBot(info.BotID)
				if err != nil {
					log.Error().Str("name", "b站推送").Err(err).Msg("")
				}
				switch {
				case info.GroupID > 0:
					if res := bdb.getAtAll(info.GroupID); res == 1 {
						msg = append([]message.MessageSegment{message.AtAll()}, msg...)
					}
					bot.SendGroupMessage(info.GroupID, msg)
				case info.GroupID < 0:
					bot.SendPrivateMessage(-info.GroupID, msg)
				}
			}
		} else if newStatus != oldStatus {
			liveStatus[key.Int()] = newStatus
		}
		return true
	})
	return nil
}
