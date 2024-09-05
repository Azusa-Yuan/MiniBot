package transform

import (
	zero "ZeroBot"
	"strconv"
)

func BidWithuid(ctx *zero.Ctx) string {
	return strconv.FormatInt(ctx.Event.SelfID, 10) + "_uid" + strconv.FormatInt(ctx.Event.SelfID, 10)
}

func BidWithgid(ctx *zero.Ctx) string {
	return strconv.FormatInt(ctx.Event.SelfID, 10) + "_gid" + strconv.FormatInt(ctx.Event.SelfID, 10)
}

func BidWithuidCustom(bid, uid int64) string {
	return strconv.FormatInt(bid, 10) + "_uid" + strconv.FormatInt(uid, 10)
}

func BidWithgidCustom(bid, gid int64) string {
	return strconv.FormatInt(bid, 10) + "_gid" + strconv.FormatInt(gid, 10)
}

func BidWithgidInt64(ctx *zero.Ctx) int64 {
	return BidWithidInt64Custom(ctx.Event.SelfID, ctx.Event.GroupID)
}

func BidWithuidInt64(ctx *zero.Ctx) int64 {
	return BidWithidInt64Custom(ctx.Event.SelfID, ctx.Event.UserID)
}

func BidWithidInt64Custom(bid, id int64) int64 {
	return (bid << 32) | id
}

func GetBidAndId(combined int64) (int64, int64) {
	bid := combined >> 32       // 右移32位获取bid
	id := combined & 0xFFFFFFFF // 取最低32位获取id
	return bid, id
}
