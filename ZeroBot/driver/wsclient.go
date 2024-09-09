package driver

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RomiChan/websocket"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	zero "ZeroBot"
	"ZeroBot/utils"
)

var (
	nullResponse = zero.APIResponse{}
)

// WSClient ...
type WSClient struct {
	seq         uint64
	conn        *websocket.Conn
	mu          sync.Mutex // 写锁
	seqMap      seqSyncMap
	Url         string // ws连接地址
	AccessToken string
	selfID      int64 // onebot的编号
}

// NewWebSocketClient 默认Driver，使用正向WS通信
func NewWebSocketClient(url, accessToken string) *WSClient {
	return &WSClient{
		Url:         url,
		AccessToken: accessToken,
	}
}

// Connect 连接ws服务端
func (ws *WSClient) Connect() {

	log.Info().Str("name", "ws").Msgf("开始尝试连接到Websocket服务器: %v", ws.Url)
	header := http.Header{
		"X-Client-Role": []string{"Universal"},
		"User-Agent":    []string{"MiniBot/1.3.3"},
	}
	if ws.AccessToken != "" {
		header["Authorization"] = []string{"Bearer " + ws.AccessToken}
	}

	network, address := resolveURI(ws.Url)
	dialer := websocket.Dialer{
		NetDial: func(_, addr string) (net.Conn, error) {
			if network == "unix" {
				host, _, err := net.SplitHostPort(addr)
				if err != nil {
					host = addr
				}
				filepath, err := base64.RawURLEncoding.DecodeString(host)
				if err == nil {
					addr = utils.BytesToString(filepath)
				}
			}
			return net.Dial(network, addr) // support unix socket transport
		},
	}

	for {
		conn, res, err := dialer.Dial(address, header)
		if err != nil {
			log.Warn().Str("name", "ws").Msgf("连接到Websocket服务器 %v 时出现错误: %v", ws.Url, err)
			time.Sleep(2 * time.Second) // 等待两秒后重新连接
			continue
		}
		ws.conn = conn
		_ = res.Body.Close()
		var rsp struct {
			SelfID int64 `json:"self_id"`
		}
		err = ws.conn.ReadJSON(&rsp)
		if err != nil {
			log.Warn().Str("name", "ws").Msgf("与Websocket服务器 %v 握手时出现错误: %v", ws.Url, err)
			time.Sleep(2 * time.Second) // 等待两秒后重新连接
			continue
		}
		ws.selfID = rsp.SelfID
		zero.APICallers.Store(ws.selfID, ws) // 添加Caller到 APICaller list...
		log.Info().Str("name", "ws").Msgf("连接Websocket服务器: %s 成功, 账号: %d", ws.Url, rsp.SelfID)
		break
	}
}

// Listen 开始监听事件
func (ws *WSClient) Listen(handler func([]byte, zero.APICaller)) {
	for {
		t, payload, err := ws.conn.ReadMessage()
		if err != nil { // reconnect
			zero.APICallers.Delete(ws.selfID) // 断开从apicaller中删除
			log.Warn().Str("name", "ws").Msg("Websocket服务器连接断开...")
			time.Sleep(time.Millisecond * time.Duration(3))
			ws.Connect()
			continue
		}
		if t != websocket.TextMessage {
			continue
		}
		rsp := gjson.Parse(utils.BytesToString(payload))
		if rsp.Get("echo").Exists() { // 存在echo字段，是api调用的返回
			log.Debug().Str("name", "ws").Msgf("接收到API调用返回: %v", strings.TrimSpace(utils.BytesToString(payload)))
			// 这个管道配合下面的Store来,判断echo是否有响应
			if c, ok := ws.seqMap.LoadAndDelete(rsp.Get("echo").Uint()); ok {
				c <- zero.APIResponse{ // 发送api调用响应
					Status:  rsp.Get("status").String(),
					Data:    rsp.Get("data"),
					Msg:     rsp.Get("msg").Str,
					Wording: rsp.Get("wording").Str,
					RetCode: rsp.Get("retcode").Int(),
					Echo:    rsp.Get("echo").Uint(),
				}
				close(c) // channel only use once
			}
			continue
		}
		if rsp.Get("meta_event_type").Str == "heartbeat" { // 忽略心跳事件
			continue
		}
		log.Debug().Str("name", "ws").Msgf("接收到事件:  %v", utils.BytesToString(payload))

		handler(payload, ws)
	}
}

func (ws *WSClient) nextSeq() uint64 {
	return atomic.AddUint64(&ws.seq, 1)
}

// CallApi 发送ws请求
func (ws *WSClient) CallApi(req zero.APIRequest) (zero.APIResponse, error) {
	ch := make(chan zero.APIResponse, 1)
	req.Echo = ws.nextSeq()
	ws.seqMap.Store(req.Echo, ch)

	data, _ := json.Marshal(&req)
	// send message
	ws.mu.Lock() // websocket write is not goroutine safe
	err := ws.conn.WriteMessage(websocket.BinaryMessage, data)
	ws.mu.Unlock()
	if err != nil {
		log.Warn().Str("name", "ws").Err(err).Msg("向WebsocketServer发送API请求失败: ")

		return nullResponse, err
	}
	log.Debug().Str("name", "ws").Msgf("向服务器发送请求: %v", utils.BytesToString(data))

	select { // 等待数据返回
	case rsp, ok := <-ch:
		if !ok {
			return nullResponse, io.ErrClosedPipe
		}
		return rsp, nil
	case <-time.After(time.Minute):
		return nullResponse, os.ErrDeadlineExceeded
	}
}
