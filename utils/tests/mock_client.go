package tests

import (
	"MiniBot/config"
	"MiniBot/utils/path"
	"os"
	"path/filepath"
	"time"

	"ZeroBot/message"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

var respChan = make(chan gjson.Result, 10)
var utilsName = "mock_client"

type MockClient struct {
	conn   *websocket.Conn
	header map[string]interface{}
}

var DefaultHeader = map[string]interface{}{
	"self_id":      10001000,
	"user_id":      123456,
	"time":         1723981165,
	"message_id":   int64(12),
	"message_type": "private",

	"font":           14,
	"sub_type":       "friend",
	"message_format": "array",
	"post_type":      "message",
}

func getDefaultHeader() map[string]interface{} {
	header := map[string]interface{}{}
	for k, v := range DefaultHeader {
		header[k] = v
	}
	return header
}

func CreatMockClient() MockClient {
	// 连接到 WebSocket 服务器
	pwdPath := path.PWDPath
	configPath := "config.yaml"
	if os.Getenv("ENVIRONMENT") == "dev" {
		configPath = "config_dev.yaml"
		log.Info().Str("name", utilsName).Msg("目前处于开发环境，请注意主目录下所有配置文件的内容和路径是否正确")
	}
	data, err := os.ReadFile(filepath.Join(pwdPath, configPath))
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}
	conf := config.MiniConfig{}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	wss := conf.WSS[0]
	conn, _, err := websocket.DefaultDialer.Dial(wss.URL+"/MiniBot", nil)
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}

	conn.WriteJSON(map[string]interface{}{"self_id": int64(123456)})
	// 启动协程接收消息
	go func() {
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				log.Error().Str("name", utilsName).Err(err).Msg("")
				return
			}
			resp := gjson.ParseBytes(payload)
			log.Info().Str("name", utilsName).Msgf("Received: %v\n", resp.Raw)

			echo := resp.Get("echo").Uint()
			if echo == 0 {
				continue
			}
			echoMsg := map[string]uint64{"echo": echo}
			conn.WriteJSON(echoMsg)

			// 过滤掉标记已回的消息
			if resp.Get("action").String() == "mark_msg_as_read" {
				continue
			}
			respChan <- resp
		}
	}()

	mockClient := MockClient{conn: conn, header: getDefaultHeader()}
	return mockClient
}

// 快速发送普通私聊消息
func (c *MockClient) Send(msg string) {
	header := c.header
	header["raw_message"] = msg
	messageArray := message.Message{}
	messageData := map[string]string{"text": msg}
	messageUnit := message.MessageSegment{Type: "text", Data: messageData}
	messageArray = append(messageArray, messageUnit)
	header["message"] = messageArray
	header["sender"] = map[string]interface{}{"user_id": header["user_id"], "nickname": "吃呀吃呀吃呀吃呀吃呀吃", "age": 18}

	err := c.conn.WriteJSON(&header)
	if err != nil {
		log.Fatal().Str("name", utilsName).Err(err).Msg("")
	}
	log.Debug().Str("name", utilsName).Msgf("向服务器发送请求:%v", &header)
}

// 快速获得初始子符串消息
func (c *MockClient) Get() string {
	timer := time.NewTicker(3 * time.Second)
	select {
	case <-timer.C:
		log.Fatal().Str("name", utilsName).Msg("超时")
		return ""
	case resp := <-respChan:
		res := ""
		array := resp.Get("params").Get("message").Array()
		for _, unit := range array {
			if unit.Get("type").String() == "text" {
				res += unit.Get("data").Get("text").String()
			}
		}
		return res
	}
}

func (c *MockClient) SetHeader(new map[string]interface{}) {
	for k, v := range new {
		c.header[k] = v
	}
}

func (c *MockClient) ResetHeader() {
	c.header = DefaultHeader
}

func (c *MockClient) SetUid(uid int64) {
	c.header["user_id"] = uid
}
func init() {

}
