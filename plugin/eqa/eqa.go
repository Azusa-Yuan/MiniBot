package eqa

import (
	"MiniBot/utils"
	database "MiniBot/utils/db"
	"MiniBot/utils/net_tools"
	"MiniBot/utils/path"
	zero "ZeroBot"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"ZeroBot/message"

	"github.com/rs/zerolog/log"
)

var (
	messageMap = map[string][]eqa{}
	Lock       = sync.RWMutex{}
	pluginName = "eqa"
)

func init() {
	metaData := zero.Metadata{
		Name: "eqa",
		Help: `### 例子： 设置在默认的情况下
##### 设置一个问题：
- 大家说111回答222
- 群友说111回答222`,
		Level: 100,
	}
	engine := zero.NewTemplate(&metaData)
	dataPath := path.GetPluginDataPath()

	initData()
	// 大家说只支持文本  回答支持多种格式
	engine.OnPrefixGroup([]string{"大家说", "群友说"}, zero.SuperUserPermission).Handle(
		func(ctx *zero.Ctx) {
			tmpMessage := ctx.Event.Message
			prefix := ctx.State["prefix"].(string)
			tmpMessage[0].Data["text"] = strings.TrimPrefix(tmpMessage[0].Data["text"], prefix)
			index := strings.Index(tmpMessage[0].Data["text"], "回答")
			if index == -1 {
				return
			}
			question := tmpMessage[0].Data["text"][:index]

			tmpMessage[0].Data["text"] = tmpMessage[0].Data["text"][index:]
			tmpMessage[0].Data["text"] = strings.TrimPrefix(tmpMessage[0].Data["text"], "回答")

			answer := message.Message{}

			for _, segment := range tmpMessage {
				switch segment.Type {
				case "text":
					if segment.Data["text"] != "" {
						answer = append(answer, message.Text(segment.Data["text"]))
					}
				case "image":
					// 计算hash值
					h := sha1.New()
					h.Write([]byte(segment.Data["url"]))
					hash := h.Sum(nil)
					hashString := hex.EncodeToString(hash)

					data, err := net_tools.DownloadWithoutTLSVerify(segment.Data["url"])
					if err != nil {
						ctx.SendError(err)
						return
					}
					imgPath := filepath.Join(dataPath, hashString)
					err = os.WriteFile(filepath.Join(dataPath, hashString), data, 0555)
					if err != nil {
						ctx.SendError(err)
						return
					}
					answer = append(answer, message.ImagePath(imgPath))
				}
			}

			data, err := json.MarshalIndent(answer, "", " ")
			if err != nil {
				ctx.SendError(err)
				return
			}

			var gid int64 = 0
			if prefix == "群友说" {
				gid = ctx.Event.GroupID
			}
			db := database.GetDefalutDB()
			qa := eqa{Key: question, Value: utils.BytesToString(data), GID: gid}
			err = db.Create(&qa).Error
			if err != nil {
				ctx.SendError(err)
				return
			}

			Lock.Lock()
			if messageList, ok := messageMap[question]; ok {
				messageMap[question] = append(messageList, qa)
			} else {
				messageMap[question] = []eqa{qa}
			}
			Lock.Unlock()

			msg := fmt.Sprintf("设置问题%s成功,回答为%s", question, utils.BytesToString(data))
			ctx.SendChain(message.Text(msg))
		},
	)

	engine.OnPrefixGroup([]string{"不要回答"}, zero.SuperUserPermission).Handle(
		func(ctx *zero.Ctx) {
			q := ctx.State["args"].(string)
			Lock.Lock()
			defer Lock.Unlock()
			if _, ok := messageMap[q]; ok {
				delete(messageMap, q)
				db := database.GetDefalutDB()
				db.Where("key = ?", q).Delete(&eqa{})
				ctx.SendChain(message.Text("该问题已删除"))
			} else {
				ctx.SendChain(message.Text("没有该问题"))
			}
		},
	)

	zero.GolbaleMiddleware.End(
		func(ctx *zero.Ctx) {
			if ctx.Event == nil || ctx.Event.Message == nil {
				return
			}
			Lock.RLock()
			defer Lock.RUnlock()
			if messageList, ok := messageMap[ctx.MessageString()]; ok {
				tmpMessageList := []eqa{}
				for _, qa := range messageList {
					if qa.GID == 0 || qa.GID == ctx.Event.GroupID {
						tmpMessageList = append(tmpMessageList, qa)
					}
				}
				if len(tmpMessageList) == 0 {
					return
				}
				respQA := messageList[rand.IntN(len(tmpMessageList))]
				ctx.SendChain(respQA.MessageList...)
			}
		},
	)
}

func initData() {
	db := database.GetDefalutDB()
	db.AutoMigrate(&eqa{})
	qas := []eqa{}
	err := db.Find(&qas).Error
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return
	}
	Lock.Lock()
	defer Lock.Unlock()

	for _, qa := range qas {
		msg := message.Message{}
		err := json.Unmarshal(utils.StringToBytes(qa.Value), &msg)
		qa.MessageList = msg
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
		}

		if messageList, ok := messageMap[qa.Key]; ok {
			messageMap[qa.Key] = append(messageList, qa)
		} else {
			messageMap[qa.Key] = []eqa{qa}
		}
	}
}
