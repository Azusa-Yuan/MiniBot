// Package bilibili bilibili卡片解析
package bilibili

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	zero "ZeroBot"

	"ZeroBot/message"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/web"
	"github.com/tidwall/gjson"
)

var (
	searchUrl        = `((b23|acg).tv|bili2233.cn)/[0-9a-zA-Z]+`
	searchVideo      = `bilibili.com\\?/video\\?/(?:av(\d+)|([bB][vV][0-9a-zA-Z]+))`
	searchDynamic    = `(t.bilibili.com|m.bilibili.com\\?/dynamic)\\?/(\d+)`
	searchArticle    = `bilibili.com\\?/read\\?/(?:cv|mobile\\?/)(\d+)`
	searchLiveRoom   = `live.bilibili.com\\?/(\d+)`
	searchUrlRe      = regexp.MustCompile(searchUrl)
	searchVideoRe    = regexp.MustCompile(searchVideo)
	searchDynamicRe  = regexp.MustCompile(searchDynamic)
	searchArticleRe  = regexp.MustCompile(searchArticle)
	searchLiveRoomRe = regexp.MustCompile(searchLiveRoom)
)

// 插件主体
func init() {
	en := zero.NewTemplate(&zero.Metadata{
		Name: "B站链接解析",
		// Level: 2,
	})

	en.OnMessage(bilibiliParseRule).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[0]
			realurl, err := bz.GetRealURL("https://" + url)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			switch {
			case searchVideoRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchVideoRe.FindStringSubmatch(realurl)
				handleVideo(ctx)
			case searchDynamicRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchDynamicRe.FindStringSubmatch(realurl)
				handleDynamic(ctx)
			case searchArticleRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchArticleRe.FindStringSubmatch(realurl)
				handleArticle(ctx)
			case searchLiveRoomRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchLiveRoomRe.FindStringSubmatch(realurl)
				handleLive(ctx)
			}
		})
	en.OnRegex(searchVideo).SetBlock(true).Handle(handleVideo)
	en.OnRegex(searchDynamic).SetBlock(true).Handle(handleDynamic)
	en.OnRegex(searchArticle).SetBlock(true).Handle(handleArticle)
	en.OnRegex(searchLiveRoom).SetBlock(true).Handle(handleLive)
}

func bilibiliParseRule(ctx *zero.Ctx) bool {
	// 消息正则匹配
	msg := ctx.MessageString()
	if matched := searchUrlRe.FindStringSubmatch(msg); matched != nil {
		ctx.State["regex_matched"] = matched
		return true
	}
	// 解析小程序
	for _, message := range ctx.Event.Message {
		if message.Type == "json" {
			data := message.Data["data"]
			datajson := gjson.Parse(data)
			url := datajson.Get("meta").Get("detail_1").Get("qqdocurl").String()
			if matched := searchUrlRe.FindStringSubmatch(url); matched != nil {
				ctx.State["regex_matched"] = matched
				return true
			}
		}
	}

	return false
}

func handleVideo(ctx *zero.Ctx) {
	id := ctx.State["regex_matched"].([]string)[1]
	if id == "" {
		id = ctx.State["regex_matched"].([]string)[2]
	}
	card, err := bz.GetVideoInfo(id)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	msg, err := videoCard2msg(card)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	summaryMsg, err := getVideoSummary(cfg, card)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	msg = append(msg, summaryMsg...)
	ctx.SendChain(msg...)
}

func handleDynamic(ctx *zero.Ctx) {
	msg, err := dynamicDetail(cfg, ctx.State["regex_matched"].([]string)[2])
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(msg...)
}

func handleArticle(ctx *zero.Ctx) {
	card, err := bz.GetArticleInfo(ctx.State["regex_matched"].([]string)[1])
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(articleCard2msg(card, ctx.State["regex_matched"].([]string)[1])...)
}

func handleLive(ctx *zero.Ctx) {
	card, err := bz.GetLiveRoomInfo(ctx.State["regex_matched"].([]string)[1])
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(liveCard2msg(card)...)
}

// getVideoSummary AI视频总结
func getVideoSummary(cookiecfg *bz.CookieConfig, card bz.Card) (msg []message.MessageSegment, err error) {
	var (
		data         []byte
		videoSummary bz.VideoSummary
	)
	data, err = web.RequestDataWithHeaders(web.NewDefaultClient(), bz.SignURL(fmt.Sprintf(bz.VideoSummaryURL, card.BvID, card.CID, card.Owner.Mid)), "GET", func(req *http.Request) error {
		if cookiecfg != nil {
			cookie := ""
			cookie, err = cookiecfg.Load()
			if err != nil {
				return err
			}
			req.Header.Add("cookie", cookie)
		}
		req.Header.Set("User-Agent", ua)
		return nil
	}, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &videoSummary)
	msg = make([]message.MessageSegment, 0, 16)
	msg = append(msg, message.Text("已为你生成视频总结\n\n"))
	msg = append(msg, message.Text(videoSummary.Data.ModelResult.Summary, "\n\n"))
	for _, v := range videoSummary.Data.ModelResult.Outline {
		msg = append(msg, message.Text("● ", v.Title, "\n"))
		for _, p := range v.PartOutline {
			msg = append(msg, message.Text(fmt.Sprintf("%d:%d %s\n", p.Timestamp/60, p.Timestamp%60, p.Content)))
		}
		msg = append(msg, message.Text("\n"))
	}
	return
}
