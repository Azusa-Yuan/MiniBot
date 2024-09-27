package manager

import (
	zero "ZeroBot"
	"time"
)

func init() {
	// 暂支持文本
	zero.OnPrefix("广播", zero.SuperUserPermission).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			rawGroupList, err := ctx.GetGroupList()
			if err != nil {
				ctx.SendError(err)
				return
			}
			tmpMessage := ctx.Event.Message
			for _, groupInfo := range rawGroupList.Array() {
				gid := groupInfo.Get("group_id").Int()
				ctx.SendGroupMessage(gid, tmpMessage)
				time.Sleep(5 * time.Second)
			}
		},
	)
}
