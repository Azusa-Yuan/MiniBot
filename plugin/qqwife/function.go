package qqwife

import (
	"fmt"
	"strconv"

	zero "ZeroBot"

	"ZeroBot/message"
)

var sendtext = [...][]string{
	{ // 表白成功
		"是个勇敢的孩子(*/ω＼*) 今天的运气都降临在你的身边~\n\n",
		"(´･ω･`)对方答应了你 并表示愿意当今天的CP\n\n",
	},
	{ // 表白失败
		"今天的运气有一点背哦~明天再试试叭",
		"_(:з」∠)_下次还有机会 咱抱抱你w",
		"今天失败了惹. 摸摸头~咱明天还有机会",
	},
	{ // ntr成功
		"因为你的个人魅力~~今天他就是你的了w\n\n",
	},
	{ // 离婚失败
		"打是情,骂是爱,不打不亲不相爱。答应我不要分手。",
		"床头打架床尾和，夫妻没有隔夜仇。安啦安啦，不要闹变扭。",
	},
	{ // 离婚成功
		"离婚成功力\n话说你不考虑当个1？",
		"离婚成功力\n天涯何处无芳草，何必单恋一枝花？不如再摘一支（bushi",
	},
}

// 注入判断 是否单身条件
func checkSingleDog(gid int64, uid int64, fiancee int64) (res bool, msg string) {

	// 判断是否需要重置
	err := qqwife.IfToday()
	if err != nil {
		msg = fmt.Sprint("[ERROR]:", err)
		return
	}
	dbinfo, err := qqwife.GetGroupInfo(gid)
	if err != nil {
		msg = fmt.Sprint("[ERROR]:", err)
		return
	}
	// 判断是否符合条件
	if dbinfo.CanMatch == 0 {
		msg = "你群包分配,别在娶妻上面下功夫，好好水群"
		return
	}
	// JudgeCD
	ok, err := qqwife.JudgeCD(gid, uid, "嫁娶", dbinfo.CDtime)

	if err != nil {
		msg = fmt.Sprint("[ERROR]:", err)
		return
	}

	if !ok {
		msg = "你的技能还在CD中..."
		return
	}

	qqwife.RLock()
	defer qqwife.RUnlock()
	// 获取用户信息
	userInfo, err := qqwife.GetMarriageInfo(gid, uid)
	if err != nil {
		msg = fmt.Sprint("[ERROR]:", err)
		return
	}
	switch {
	case userInfo != (UserInfo{}) && (userInfo.Target == 0 || userInfo.UID == 0): // 如果是单身贵族
		msg = "今天的你是单身贵族噢"
		return
	case userInfo.Target == fiancee:
		msg = "笨蛋！你们已经在一起了！"
		return
	case userInfo.Mode == 1: // 如果如为攻
		msg = "笨蛋~你家里还有个吃白饭的w"
		return
	case userInfo != (UserInfo{}) && userInfo.Mode == 0: // 如果为受
		msg = "该是0就是0,当0有什么不好"
		return
	}
	fianceeInfo, _ := qqwife.GetMarriageInfo(gid, fiancee)
	switch {
	case fianceeInfo != (UserInfo{}) && (fianceeInfo.Target == 0 || fianceeInfo.UID == 0): // 如果是单身贵族
		msg = "今天的ta是单身贵族噢"
		return
	case fianceeInfo.Mode == 1: // 如果如为攻
		msg = "他有别的女人了，你该放下了"
		return
	case fianceeInfo != (UserInfo{}) && fianceeInfo.Mode == 0: // 如果为受
		msg = "ta被别人娶了,你来晚力"
		return
	}
	res = true
	return
}

// 注入判断 是否满足小三要求
func checkMistress(ctx *zero.Ctx) (ok bool, targetInfo UserInfo) {

	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	fid := ctx.State["regex_matched"].([]string)
	target, err := strconv.ParseInt(fid[3]+fid[4], 10, 64)
	if err != nil {
		target, err = strconv.ParseInt(fid[8]+fid[9], 10, 64)
	}
	if err != nil {
		ctx.SendChain(message.Text("额,你的target好像不存在?"))
		return
	}
	// 判断是否需要重置
	err = qqwife.IfToday()
	if err != nil {
		ctx.SendChain(message.Text("[ERROR]:", err))
		return
	}
	groupInfo, err := qqwife.GetGroupInfo(gid)
	if err != nil {
		ctx.SendChain(message.Text("[ERROR]:", err))
		return
	}
	if groupInfo.CanNtr == 0 {
		ctx.SendChain(message.Text("你群发布了牛头人禁止令，放弃吧"))
		return
	}
	// JudgeCD
	cdOK, err := qqwife.JudgeCD(gid, uid, "NTR", groupInfo.CDtime)
	switch {
	case err != nil:
		ctx.SendChain(message.Text("[ERROR]:", err))
		return
	case !cdOK:
		ctx.SendChain(message.Text("你的技能还在CD中..."))
		return
	}
	// 获取用户信息
	qqwife.RLock()
	defer qqwife.RUnlock()
	targetInfo, _ = qqwife.GetMarriageInfo(gid, target)
	switch {
	case targetInfo == (UserInfo{}): // 如果是空数据
		ctx.SendChain(message.Text("ta现在还是单身哦,快向ta表白吧!"))
		return
	case targetInfo.Target == 0 || targetInfo.UID == 0: // 如果是单身贵族
		ctx.SendChain(message.Text("今天的ta是单身贵族噢"))
		return
	case targetInfo.Target == uid || targetInfo.UID == uid:
		ctx.SendChain(message.Text("笨蛋！你们已经在一起了！"))
		return
	}
	// 获取用户信息
	userInfo, _ := qqwife.GetMarriageInfo(gid, uid)
	switch {
	case userInfo != (UserInfo{}) && (userInfo.Target == 0 || userInfo.UID == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的你是单身贵族噢"))
		return
	case userInfo.Mode == 1: // 如果如为攻
		ctx.SendChain(message.Text("打灭，不给纳小妾！"))
		return
	case userInfo != (UserInfo{}) && userInfo.Mode == 0: // 如果为受
		ctx.SendChain(message.Text("该是0就是0,当0有什么不好"))
		return
	}
	ok = true
	return
}

func checkDivorce(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	// 判断是否需要重置
	err := qqwife.IfToday()
	if err != nil {
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	}
	groupInfo, err := qqwife.GetGroupInfo(gid)
	if err != nil {
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	}
	ok, err := qqwife.JudgeCD(gid, uid, "离婚", groupInfo.CDtime)
	switch {
	case err != nil:
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	case !ok:
		ctx.SendChain(message.Text("你的技能还在CD中..."))
		return false
	}
	return true
}

func checkMatchmaker(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	gayOne, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)

	if err != nil {
		ctx.SendChain(message.Text("额，攻方好像不存在？"))
		return false
	}
	gayZero, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，受方好像不存在？"))
		return false
	}
	selfId := ctx.Event.SelfID
	// 判断是否机器人自身
	if selfId == gayOne || selfId == gayZero {
		return false
	}
	if gayOne == uid || gayZero == uid {
		ctx.SendChain(message.Text("禁止自己给自己做媒!"))
		return false
	}
	if gayOne == gayZero {
		ctx.SendChain(message.Text("你这个媒人XP很怪咧,不能这样噢"))
		return false
	}
	// 判断是否需要重置
	err = qqwife.IfToday()
	if err != nil {
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	}
	groupInfo, err := qqwife.GetGroupInfo(gid)
	if err != nil {
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	}
	ok, err := qqwife.JudgeCD(gid, uid, "做媒", groupInfo.CDtime)
	switch {
	case err != nil:
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	case !ok:
		ctx.SendChain(message.Text("你的技能还在CD中..."))
		return false
	}
	qqwife.RLock()
	defer qqwife.RUnlock()
	gayOneInfo, _ := qqwife.GetMarriageInfo(gid, gayOne)
	switch {
	case gayOneInfo != (UserInfo{}) && (gayOneInfo.Target == 0 || gayOneInfo.UID == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的攻方是单身贵族噢"))
		return false
	case gayOneInfo.Target == gayZero || gayOneInfo.UID == gayZero:
		ctx.SendChain(message.Text("笨蛋!ta们已经在一起了!"))
		return false
	case gayOneInfo != (UserInfo{}): // 如果不是单身
		ctx.SendChain(message.Text("攻方不是单身,不允许给这种人做媒!"))
		return false
	}
	// 获取用户信息
	gayZeroInfo, _ := qqwife.GetMarriageInfo(gid, gayZero)
	switch {
	case gayOneInfo != (UserInfo{}) && (gayZeroInfo.Target == 0 || gayZeroInfo.UID == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的攻方是单身贵族噢"))
		return false
	case gayZeroInfo != (UserInfo{}): // 如果不是单身
		ctx.SendChain(message.Text("受方不是单身,不允许给这种人做媒!"))
		return false
	}
	return true
}
