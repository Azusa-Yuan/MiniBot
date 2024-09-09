package zero

import (
	"fmt"

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
		} else {
			log.Info().Str("name", "bot").Msgf("%v在群%v戳了戳%v", e.UserID, e.GroupID, e.TargetID)
		}
	case e.NoticeType == "group_card":
		log.Info().Str("name", "bot").Msgf("%v在群%v修改了群名片 %v → %v", e.UserID, e.GroupID, e.CardOld, e.CardNew)

	}
	fmt.Println(e.RawEvent.Raw)
}
