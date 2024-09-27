package zero

import (
	"hash/crc64"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ZeroBot/message"
	"ZeroBot/utils"
)

// Type check the ctx.Event's type PostType DetailType SubType
func Type(type_ string) Rule {
	t := strings.SplitN(type_, "/", 3)
	return func(ctx *Ctx) bool {
		if len(t) > 0 && t[0] != ctx.Event.PostType {
			return false
		}
		if len(t) > 1 && t[1] != ctx.Event.DetailType {
			return false
		}
		if len(t) > 2 && t[2] != ctx.Event.SubType {
			return false
		}
		return true
	}
}

// PrefixRule check if the message has the prefix and trim the prefix
//
// 检查消息前缀
//
//	func PrefixRule(prefixes ...string) Rule {
//		return func(ctx *Ctx) bool {
//			if len(ctx.Event.Message) == 0 || ctx.Event.Message[0].Type != "text" { // 确保无空指针
//				return false
//			}
//			first := ctx.Event.Message[0]
//			firstMessage := first.Data["text"]
//			for _, prefix := range prefixes {
//				if strings.HasPrefix(firstMessage, prefix) {
//					ctx.State["prefix"] = prefix
//					arg := strings.TrimLeft(firstMessage[len(prefix):], " ")
//					if len(ctx.Event.Message) > 1 {
//						arg += ctx.Event.Message[1:].ExtractPlainText()
//					}
//					ctx.State["args"] = arg
//					return true
//				}
//			}
//			return false
//		}
//	}
//
// 检查消息前缀
func PrefixRule(prefixes ...string) Rule {
	return func(ctx *Ctx) bool {
		if len(ctx.Event.Message) == 0 { // 确保无空指针
			return false
		}
		for _, msg := range ctx.Event.Message {
			if msg.Type == "text" {
				// 适配lag
				if strings.TrimSpace(msg.Data["text"]) == "" {
					continue
				}
				firstMessage := strings.TrimLeft(msg.Data["text"], " ")
				for _, prefix := range prefixes {
					if strings.HasPrefix(firstMessage, prefix) {
						extractPlainText := ctx.ExtractPlainText()
						ctx.State["prefix"] = prefix
						arg := strings.TrimLeft(extractPlainText[len(prefix):], " ")
						ctx.State["args"] = arg
						return true
					}
				}
				// 只匹配第一个为text类型的的非空message
				break
			}
		}
		return false
	}
}

// SuffixRule check if the message has the suffix and trim the suffix
//
// 检查消息后缀
func SuffixRule(suffixes ...string) Rule {
	return func(ctx *Ctx) bool {
		mLen := len(ctx.Event.Message)
		if mLen <= 0 { // 确保无空指针
			return false
		}
		last := ctx.Event.Message[mLen-1]
		if last.Type != "text" {
			return false
		}
		lastMessage := last.Data["text"]
		for _, suffix := range suffixes {
			if strings.HasSuffix(lastMessage, suffix) {
				ctx.State["suffix"] = suffix
				arg := strings.TrimRight(lastMessage[:len(lastMessage)-len(suffix)], " ")
				if mLen >= 2 {
					arg += ctx.Event.Message[:mLen].ExtractPlainText()
				}
				ctx.State["args"] = arg
				return true
			}
		}
		return false
	}
}

// CommandRule check if the message is a command and trim the command name
func CommandRule(commands ...string) Rule {
	return func(ctx *Ctx) bool {
		if len(ctx.Event.Message) == 0 || ctx.Event.Message[0].Type != "text" {
			return false
		}
		first := ctx.Event.Message[0]
		firstMessage := first.Data["text"]
		if !strings.HasPrefix(firstMessage, BotConfig.CommandPrefix) {
			return false
		}
		cmdMessage := firstMessage[len(BotConfig.CommandPrefix):]
		for _, command := range commands {
			if strings.HasPrefix(cmdMessage, command) {
				ctx.State["command"] = command
				arg := strings.TrimLeft(cmdMessage[len(command):], " ")
				if len(ctx.Event.Message) > 1 {
					arg += ctx.Event.Message[1:].ExtractPlainText()
				}
				ctx.State["args"] = arg
				return true
			}
		}
		return false
	}
}

// RegexRule check if the message can be matched by the regex pattern
func RegexRule(regexPattern string) Rule {
	regex := regexp.MustCompile(regexPattern)
	return func(ctx *Ctx) bool {
		msg := ctx.MessageString()
		if matched := regex.FindStringSubmatch(msg); matched != nil {
			ctx.State["regex_matched"] = matched
			return true
		}
		return false
	}
}

// ReplyRule check if the message is replying some message
func ReplyRule(messageID int64) Rule {
	return func(ctx *Ctx) bool {
		if len(ctx.Event.Message) == 0 {
			return false
		}
		if ctx.Event.Message[0].Type != "reply" {
			return false
		}
		if id, err := strconv.ParseInt(ctx.Event.Message[0].Data["id"], 10, 64); err == nil {
			return id == messageID
		}
		c := crc64.New(crc64.MakeTable(crc64.ISO))
		c.Write(utils.StringToBytes(ctx.Event.Message[0].Data["id"]))
		return int64(c.Sum64()) == messageID
	}
}

// KeywordRule check if the message has a keyword or keywords 用ExtractPlainText防止图片base64触发
func KeywordRule(src ...string) Rule {
	return func(ctx *Ctx) bool {
		msg := ctx.ExtractPlainText()
		for _, str := range src {
			if strings.Contains(msg, str) {
				ctx.State["keyword"] = str
				return true
			}
		}
		return false
	}
}

// FullMatchRule check if src has the same copy of the message
func FullMatchRule(src ...string) Rule {
	return func(ctx *Ctx) bool {
		msg := ctx.MessageString()
		for _, str := range src {
			if str == msg {
				ctx.State["matched"] = msg
				return true
			}
		}
		return false
	}
}

// OnlyToMe only triggered in conditions of @bot or begin with the nicknames
func OnlyToMe(ctx *Ctx) bool {
	return ctx.Event.IsToMe
}

// CheckUser only triggered by specific person
func CheckUser(userId ...int64) Rule {
	return func(ctx *Ctx) bool {
		for _, uid := range userId {
			if ctx.Event.UserID == uid {
				return true
			}
		}
		return false
	}
}

// CheckGroup only triggered in specific group
func CheckGroup(grpId ...int64) Rule {
	return func(ctx *Ctx) bool {
		for _, gid := range grpId {
			if ctx.Event.GroupID == gid {
				return true
			}
		}
		return false
	}
}

// OnlyPrivate requires that the ctx.Event is private message
func OnlyPrivate(ctx *Ctx) bool {
	return ctx.Event.PostType == "message" && ctx.Event.DetailType == "private"
}

// OnlyPublic requires that the ctx.Event is public/group or public/guild message
func OnlyPublic(ctx *Ctx) bool {
	return ctx.Event.PostType == "message" && (ctx.Event.DetailType == "group" || ctx.Event.DetailType == "guild")
}

// OnlyGroup requires that the ctx.Event is public/group message
func OnlyGroup(ctx *Ctx) bool {
	return ctx.Event.PostType == "message" && ctx.Event.DetailType == "group"
}

// OnlyGuild requires that the ctx.Event is public/guild message
func OnlyGuild(ctx *Ctx) bool {
	return ctx.Event.PostType == "message" && ctx.Event.DetailType == "guild"
}

func issu(id int64) bool {
	for _, su := range BotConfig.SuperUsers {
		if su == id {
			return true
		}
	}
	return false
}

// SuperUserPermission only triggered by the bot's owner
func SuperUserPermission(ctx *Ctx) bool {
	return issu(ctx.Event.UserID)
}

// AdminPermission only triggered by the group admins or higher permission
func AdminPermission(ctx *Ctx) bool {
	return SuperUserPermission(ctx) || ctx.Event.Sender.Role == "owner" || ctx.Event.Sender.Role == "admin"
}

// OwnerPermission only triggered by the group owner or higher permission
func OwnerPermission(ctx *Ctx) bool {
	return SuperUserPermission(ctx) || ctx.Event.Sender.Role == "owner"
}

// UserOrGrpAdmin 允许用户单独使用或群管使用
func UserOrGrpAdmin(ctx *Ctx) bool {
	if OnlyGroup(ctx) {
		return AdminPermission(ctx)
	}
	return OnlyToMe(ctx)
}

// GroupHigherPermission 群发送者权限高于 target
//
// 隐含 OnlyGroup 判断
func GroupHigherPermission(gettarget func(ctx *Ctx) int64) Rule {
	return func(ctx *Ctx) bool {
		if !OnlyGroup(ctx) {
			return false
		}
		target := gettarget(ctx)
		if target == ctx.Event.UserID { // 特判, 自己和自己比
			return false
		}
		if SuperUserPermission(ctx) {
			sender := ctx.Event.UserID
			return BotConfig.GetFirstSuperUser(sender, target) == sender
		}
		if target == ctx.Event.UserID { // 特判, 自己和自己比
			return false
		}
		if ctx.Event.Sender.Role == "owner" {
			info, err := ctx.GetThisGroupMemberInfo(target, false)
			if err != nil {
				ctx.SendError(err)
				return false
			}
			return !issu(target) && info.Get("role").Str != "owner"
		}
		if ctx.Event.Sender.Role == "admin" {
			info, err := ctx.GetThisGroupMemberInfo(target, false)
			if err != nil {
				ctx.SendError(err)
				return false
			}
			tgtrole := info.Get("role").Str
			return !issu(target) && tgtrole != "owner" && tgtrole != "admin"
		}
		return false // member is the lowest
	}
}

// HasPicture 消息含有图片返回 true
func HasPicture(ctx *Ctx) bool {
	var urls = []string{}
	for _, elem := range ctx.Event.Message {
		if elem.Type == "image" {
			if elem.Data["url"] != "" {
				urls = append(urls, elem.Data["url"])
			}
		}
	}
	if len(urls) > 0 {
		ctx.State["image_url"] = urls
		return true
	}
	return false
}

// MustProvidePicture 消息不存在图片阻塞120秒至有图片，超时返回 false
func MustProvidePicture(ctx *Ctx) bool {
	if HasPicture(ctx) {
		return true
	}
	// 没有图片就索取
	ctx.SendChain(message.Text("请发送一张图片"))
	next := NewFutureEvent("message", 999, true, ctx.CheckSession(), HasPicture).Next()
	select {
	case <-time.After(time.Second * 120):
		return false
	case newCtx := <-next:
		ctx.State["image_url"] = newCtx.State["image_url"]
		ctx.Event.MessageID = newCtx.Event.MessageID
		return true
	}
}
