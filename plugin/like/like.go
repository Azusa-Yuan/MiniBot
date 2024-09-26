package like

import (
	"MiniBot/service/book"
	zero "ZeroBot"
	"ZeroBot/message"
	"time"

	"github.com/fumiama/cron"
	"github.com/rs/zerolog/log"
)

var (
	pluginName   = "like"
	defaultTimes = 10
)

func init() {
	metaData := &zero.MetaData{
		Name: pluginName,
		Help: "点赞,艾特机器人发送 赞我 。非机器人好友有概率失败",
	}
	engine := zero.NewTemplate(metaData)
	engine.OnFullMatch("赞我", zero.OnlyToMe).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			err := ctx.SendLike(ctx.Event.UserID, defaultTimes)
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("好啦好啦,已经点赞10次啦"))
		},
	)

	engine.OnFullMatch("订阅点赞", zero.OnlyPrivate).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			bookInfo := &book.Book{
				BotID:   ctx.Event.SelfID,
				UserID:  ctx.Event.UserID,
				GroupID: ctx.Event.GroupID,
				Service: pluginName,
			}
			err := book.CreatOrUpdateBookInfo(bookInfo)
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("点赞订阅成功"))
		},
	)

	timeZone, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(cron.WithLocation(timeZone))
	c.AddFunc("25 0 * * *", SendLike)
	c.Start()
}

func SendLike() {
	bookInfos, err := book.GetBookInfos(pluginName)
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return
	}
	for _, bookInfo := range bookInfos {
		bot, err := zero.GetBot(bookInfo.BotID)
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
			continue
		}
		err = bot.SendLike(bookInfo.UserID, defaultTimes)
		if err != nil {
			bot.SendError(err)
			continue
		}
		bot.SendChain(message.Text("今天的点赞成功啦"))
	}
}
