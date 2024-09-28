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
	messageMap map[string][]message.Message
	Lock       = sync.RWMutex{}
	pluginName = "eqa"
)

func init() {
	metaData := zero.MetaData{
		Name: "eqa",
		Help: `### 例子： 设置在默认的情况下
##### 设置一个问题：
- 大家说111回答222
- 我说333回答444
- 大家说@某人回答图1图2 文字
- 大家说图片回答图片
- 有人说R测试回答test`,
		Level: 100,
	}
	engine := zero.NewTemplate(&metaData)
	dataPath := path.GetPluginDataPath()

	initData()
	// 大家说只支持文本  回答支持多种格式
	engine.OnRegex("^大家说.+回答.+", zero.SuperUserPermission).Handle(
		func(ctx *zero.Ctx) {
			// 测试用
			tmpMessage := ctx.Event.Message
			tmpMessage[0].Data["text"] = strings.TrimPrefix(tmpMessage[0].Data["text"], "大家说")
			index := strings.Index(tmpMessage[0].Data["text"], "回答")
			question := tmpMessage[0].Data["text"][:index]

			tmpMessage[0].Data["text"] = tmpMessage[0].Data["text"][index+len("回答")+1:]

			answer := message.Message{}

			for _, segment := range tmpMessage {

				switch segment.Type {
				case "text":
					answer = append(answer, message.Text(segment.Data["text"]))
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

			db := database.GetDefalutDB()
			err = db.Create(&eqa{Key: question, Value: utils.BytesToString(data)}).Error
			if err != nil {
				ctx.SendError(err)
				return
			}

			Lock.Lock()
			if messageList, ok := messageMap[question]; ok {
				messageMap[question] = append(messageList, answer)
			} else {
				messageMap[question] = []message.Message{answer}
			}
			Lock.Unlock()

			msg := fmt.Sprintf("设置问题%s成功,回答为%s", question, utils.BytesToString(data))
			ctx.SendChain(message.Text(msg))
		},
	)

	zero.GolbaleMiddleware.End(
		func(ctx *zero.Ctx) {
			if ctx.Event == nil || ctx.Event.Message == nil {
				return
			}
			Lock.RLock()
			defer Lock.RUnlock()
			resp := message.Message{}
			if messageList, ok := messageMap[ctx.MessageString()]; ok {
				resp = messageList[rand.IntN(len(messageList))]
			}
			ctx.SendChain(resp...)
		},
	)
}

func initData() {
	db := database.GetDefalutDB()
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
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
		}

		if messageList, ok := messageMap[qa.Key]; ok {
			messageMap[qa.Key] = append(messageList, msg)
		} else {
			messageMap[qa.Key] = []message.Message{msg}
		}
	}
}
