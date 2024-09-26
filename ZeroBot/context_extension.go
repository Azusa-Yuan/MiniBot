package zero

import (
	"strconv"
	"unicode"

	"github.com/rs/zerolog/log"
)

// isQQ checks if the given string is a valid QQ number and returns it as an integer.
func IsQQ(msg string) int64 {
	if len(msg) < 5 || len(msg) > 11 {
		return -1
	}
	for _, r := range msg {
		if !unicode.IsDigit(r) {
			return -1
		}
	}
	// Convert the valid QQ string to an integer
	qqNumber, err := strconv.ParseInt(msg, 10, 64)
	if err != nil {
		return -1 // Return -1 if conversion fails
	}
	return qqNumber
}

type AtInfo struct {
	QQ       int64
	NickName string
}

func (ctx *Ctx) GetAtInfos() []AtInfo {
	atInfos := []AtInfo{}
	for _, segment := range ctx.Event.Message {
		if segment.Type == "at" {
			qqStr := segment.Data["qq"]
			qq, err := strconv.ParseInt(qqStr, 10, 64)
			if err != nil {
				log.Error().Str("name", "zero").Err(err).Msg("")
				continue
			}
			atInfos = append(atInfos, AtInfo{
				QQ:       qq,
				NickName: segment.Data["name"],
			})
			continue
		}
		if segment.Type == "text" {
			qq := IsQQ(segment.Data["text"])
			if qq > 0 {
				atInfos = append(atInfos, AtInfo{
					QQ: qq,
				})
			}
		}
	}
	return atInfos
}
