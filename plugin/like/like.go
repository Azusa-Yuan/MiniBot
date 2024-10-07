package like

import (
	"MiniBot/service/book"
	"MiniBot/utils/schedule"
	zero "ZeroBot"
	"ZeroBot/message"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	pluginName   = "like"
	defaultTimes = 10
)

func init() {
	metaData := &zero.Metadata{
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

	schedule.Cron.AddFunc("25 0 * * *", SendLike)
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
			log.Error().Str("name", pluginName).Err(err).Msg("")
			continue
		}
		bot.SendPrivateMessage(bookInfo.UserID, message.Text("今天的点赞成功啦"))
		time.Sleep(5 * time.Second)
	}
}
