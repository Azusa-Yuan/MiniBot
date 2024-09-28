package manager

import (
	zero "ZeroBot"
	"time"
)

func init() {
	zero.OnPrefix("广播", zero.SuperUserPermission).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			rawGroupList, err := ctx.GetGroupList()
			if err != nil {
				ctx.SendError(err)
				return
			}
			for _, groupInfo := range rawGroupList.Array() {
				gid := groupInfo.Get("group_id").Int()
				ctx.SendGroupMessage(gid, ctx.ReceptionToSend())
				time.Sleep(5 * time.Second)
			}
		},
	)
}
