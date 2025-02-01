package meme

import (
	"MiniBot/utils/cache"
	"MiniBot/utils/net_tools"
	zero "ZeroBot"
	"ZeroBot/message"
	"encoding/json"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
)

var (
	pluginName = "表情包制作"
	pattern    = `--(\S+)\s+(\S+)`
	reArgs     = regexp.MustCompile(pattern)
	help       = `触发方式：“关键词 + 图片/文字”
发送 “表情详情 + 关键词” 查看表情参数和预览
目前支持的表情列表：`
)

func init() {
	err := InitMeme()
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return
	}
	metaData := zero.Metadata{
		Name: pluginName,
		Help: "发送 表情包列表 查看所有表情指令 \n发送 表情详情 xx 查看表情详细参数",
	}
	engine := zero.NewTemplate(&metaData)
	engine.OnFullMatchGroup([]string{"表情包列表", "头像表情包"}).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			data, err := GetHelp()
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.Text(help), message.ImageBytes(data))
		},
	)

	// 优先匹配前缀长的
	keys := maps.Keys(emojiMap)
	slices.SortFunc(keys, func(i, j string) int {
		return len(j) - len(i)
	})
	engine.OnPrefixGroup(keys).Handle(dealmeme)

	engine.OnPrefix("表情详情").SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			key := strings.TrimSpace(ctx.State["args"].(string))
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text(QueryEmojiInfo(key)))
		},
	)
}

func dealImgStr(imgStrs ...string) ([][]byte, error) {
	images := [][]byte{}
	for _, imgStr := range imgStrs {
		uid, err := strconv.ParseInt(imgStr, 10, 64)
		var data []byte
		if err != nil {
			data, err = net_tools.DownloadWithoutTLSVerify(imgStr)
		} else {
			data, err = cache.GetAvatar(uid)
		}
		if err != nil {
			return nil, err
		}
		images = append(images, data)
	}
	return images, nil
}

func dealmeme(ctx *zero.Ctx) {
	path := emojiMap[ctx.State["prefix"].(string)]
	imgStrs := []string{}
	args := map[string]any{}
	args["user_infos"] = []UserInfo{}

	messageWithReply := ctx.Event.Message
	if ctx.Event.Message[0].Type == "reply" {
		replyMsg := ctx.GetMessage(ctx.Event.Message[0].Data["id"])
		replyImg := message.Message{}
		for _, segment := range replyMsg.Elements {
			if segment.Type == "image" {
				replyImg = append(replyImg, segment)
			}
		}
		messageWithReply = append(replyImg, ctx.Event.Message[1:]...)
	}

	for _, segment := range messageWithReply {
		if segment.Type == "at" {
			qqStr := segment.Data["qq"]
			qq, err := strconv.ParseInt(qqStr, 10, 64)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
				continue
			}
			args["user_infos"] = append(args["user_infos"].([]UserInfo),
				UserInfo{
					Name:   ctx.CardOrNickName(qq),
					Gender: "female",
				})
			imgStrs = append(imgStrs, qqStr)
		}
		if segment.Type == "image" {
			imgStrs = append(imgStrs, segment.Data["url"])
		}
	}

	// 正则匹配所有args
	extractPlainText := ctx.State["args"].(string)
	matchs := reArgs.FindAllStringSubmatch(extractPlainText, -1)
	if len(matchs) > 0 {
		for _, match := range matchs {
			args[match[1]] = match[2]
		}
		extractPlainText = reArgs.ReplaceAllString(extractPlainText, "")
	}
	// Fields函数会将字符串按空格分割,并自动忽略连续的空格
	texts := strings.Fields(extractPlainText)

	// 做截断, 仅对img做截断更用户友好
	imgStrs, _ = truncateList(path, imgStrs, texts)

	if !fastJudge(path, imgStrs, texts) {
		imgStrs = append([]string{strconv.FormatInt(ctx.Event.UserID, 10)}, imgStrs...)
		args["user_infos"] = append(args["user_infos"].([]UserInfo),
			UserInfo{
				Name:   ctx.CardOrNickName(ctx.Event.UserID),
				Gender: "female",
			})
		if !fastJudge(path, imgStrs, texts) {
			return
		}
	}

	// 表情包占用前缀太多 匹配上了才阻止后续
	ctx.Stop()

	images, err := dealImgStr(imgStrs...)
	if err != nil {
		ctx.SendError(err)
		return
	}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		ctx.SendError(err)
		return
	}

	emojiData, err := CreateEmoji(path, images, texts, string(argsBytes))
	if err != nil {
		ctx.SendError(err)
		return
	}
	ctx.SendChain(message.At(ctx.Event.UserID), message.ImageBytes(emojiData))
}
