package zero

import (
	"github.com/rs/zerolog/log"
)

func printMessageLog(e *Event) {
	switch {
	case e.DetailType == "group":
		log.Info().Str("name", "bot").Msgf("收到群(%v)消息 %v : %v", e.GroupID, e.Sender.String(), e.RawMessage)
	case e.DetailType == "guild" && e.SubType == "channel":
		log.Info().Str("name", "bot").Msgf("收到频道(%v)(%v-%v)消息 %v : %v", e.GroupID, e.GuildID, e.ChannelID, e.Sender.String(), e.Message)
	default:
		log.Info().Str("name", "bot").Msgf("收到私聊消息 %v : %v", e.Sender.String(), e.RawMessage)
	}
}

func printNoticeLog(e *Event) {
	switch {
	case e.SubType == "poke":
		if e.GroupID == 0 {
			log.Info().Str("name", "bot").Msgf("%v戳了戳%v", e.UserID, e.TargetID)
			return
		} else {
			log.Info().Str("name", "bot").Msgf("%v在群%v戳了戳%v", e.UserID, e.GroupID, e.TargetID)
			return
		}
	case e.DetailType == "group_card":
		log.Info().Str("name", "bot").Msgf("%v在群%v修改了群名片 %v → %v", e.UserID, e.GroupID, e.RawEvent.Get("card_old").String(), e.RawEvent.Get("card_new").String())
		return
	case e.DetailType == "group_recall":
		log.Info().Str("name", "bot").Msgf("%d在群%d撤回了一条消息", e.OperatorID, e.GroupID)
		return
	case e.DetailType == "group_upload":
		log.Info().Str("name", "bot").Msgf("%d在群%d上传了文件%v", e.UserID, e.GroupID, e.RawEvent.Get("file").Get("name").String())
		return
	case e.DetailType == "group_increase":
		if e.SubType == "approve" {
			log.Info().Str("name", "bot").Msgf("%d同意%d加入群聊%d", e.OperatorID, e.UserID, e.GroupID)
		}
		return
	}
	log.Info().Str("name", "notice_test").Msgf(e.RawEvent.Raw)
}

func printRequestLog(e *Event) {
	log.Info().Str("name", "request_test").Msgf(e.RawEvent.Raw)
}
