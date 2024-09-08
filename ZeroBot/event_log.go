package zero

import "github.com/rs/zerolog/log"

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
		log.Info().Str("name", "bot").Msgf("%v戳了戳%v   %v", e.UserID, e.TargetID, e)
	default:
		log.Info().Str("name", "bot").Str("type", "post").Msgf("%v", e)
	}
}
