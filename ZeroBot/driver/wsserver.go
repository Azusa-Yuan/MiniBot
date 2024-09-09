package driver

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/RomiChan/websocket"
	"github.com/tidwall/gjson"

	zero "ZeroBot"
	"ZeroBot/utils"

	"github.com/rs/zerolog/log"
)

// WSServer ...
type WSServer struct {
	Url         string // ws连接地址
	AccessToken string
	lstn        net.Listener
	caller      chan *WSSCaller

	json.Unmarshaler
}

// UnmarshalJSON init WSServer with waitn=16
func (wss *WSServer) UnmarshalJSON(data []byte) error {
	type jsoncfg struct {
		Url         string // ws连接地址
		AccessToken string
	}
	err := json.Unmarshal(data, (*jsoncfg)(unsafe.Pointer(wss)))
	if err != nil {
		return err
	}
	wss.caller = make(chan *WSSCaller, 16)
	return nil
}

// NewWebSocketServer 使用反向WS通信
func NewWebSocketServer(waitn int, url, accessToken string) *WSServer {
	return &WSServer{
		Url:         url,
		AccessToken: accessToken,
		caller:      make(chan *WSSCaller, waitn),
	}
}

// WSSCaller ...
type WSSCaller struct {
	mu     sync.Mutex // 写锁
	seqMap seqSyncMap
	conn   *websocket.Conn
	selfID int64
	seq    uint64
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func checkAuth(req *http.Request, token string) int {
	if token == "" { // quick path
		return http.StatusOK
	}

	auth := req.Header.Get("Authorization")
	if auth == "" {
		auth = req.URL.Query().Get("access_token")
	} else {
		_, after, ok := strings.Cut(auth, " ")
		if ok {
			auth = after
		}
	}

	switch auth {
	case token:
		return http.StatusOK
	case "":
		return http.StatusUnauthorized
	default:
		return http.StatusForbidden
	}
}

func (wss *WSServer) any(w http.ResponseWriter, r *http.Request) {
	status := checkAuth(r, wss.AccessToken)
	if status != http.StatusOK {
		log.Warn().Str("name", "wss").Msgf("已拒绝 %v 的 WebSocket 请求: Token鉴权失败(code:%d)", r.RemoteAddr, status)
		w.WriteHeader(status)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn().Str("name", "wss").Msgf("处理 WebSocket 请求时出现错误: %v", err)
		return
	}

	var rsp struct {
		SelfID int64 `json:"self_id"`
	}
	err = conn.ReadJSON(&rsp)
	if err != nil {
		log.Warn().Str("name", "wss").Msgf("与Websocket服务器 %v 握手时出现错误: %v", wss.Url, err)
		return
	}

	c := &WSSCaller{
		conn:   conn,
		selfID: rsp.SelfID,
	}
	zero.APICallers.Store(rsp.SelfID, c) // 添加Caller到 APICaller list...
	log.Info().Str("name", "wss").Msgf("连接Websocket服务器: %s 成功, 账号: %d", wss.Url, rsp.SelfID)
	wss.caller <- c
}

// Connect 监听ws服务
func (wss *WSServer) Connect() {
	network, address := resolveURI(wss.Url)
	uri, err := url.Parse(address)
	if err == nil && uri.Scheme != "" {
		address = uri.Host
	}

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Warn().Str("name", "wss").Msgf("Websocket服务器监听失败:%v", err)

		wss.lstn = nil
		return
	}

	wss.lstn = listener
	log.Info().Str("name", "wss").Msgf("Websocket服务器开始监听:%v", listener.Addr())
}

// Listen 开始监听事件
func (wss *WSServer) Listen(handler func([]byte, zero.APICaller)) {
	mux := http.ServeMux{}
	mux.HandleFunc("/MiniBot", wss.any)
	go func() {
		for {
			if wss.lstn == nil {
				time.Sleep(time.Millisecond * time.Duration(3))
				wss.Connect()
				continue
			}
			log.Info().Str("name", "wss").Msgf("WebSocket 服务器开始处理: %v", wss.lstn.Addr())
			err := http.Serve(wss.lstn, &mux)
			if err != nil {
				log.Warn().Str("name", "wss").Msgf("Websocket服务器在端点%v 失败:%v", wss.lstn.Addr(), err)
				wss.lstn = nil
			}
		}
	}()

	// 这是个管道  所以会卡死在这
	for wssc := range wss.caller {
		go wssc.listen(handler)
	}
}

func (wssc *WSSCaller) listen(handler func([]byte, zero.APICaller)) {
	for {
		t, payload, err := wssc.conn.ReadMessage()
		if err != nil { // reconnect
			zero.APICallers.Delete(wssc.selfID) // 断开从apicaller中删除
			log.Warn().Str("name", "wss").Msg("Websocket服务器连接断开...")
			return
		}
		if t != websocket.TextMessage {
			continue
		}
		rsp := gjson.Parse(utils.BytesToString(payload))
		if rsp.Get("echo").Exists() { // 存在echo字段，是api调用的返回
			log.Debug().Str("name", "wss").Msgf("接收到API调用返回: %v", strings.TrimSpace(utils.BytesToString(payload)))
			if c, ok := wssc.seqMap.LoadAndDelete(rsp.Get("echo").Uint()); ok {
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
		log.Debug().Str("name", "wss").Msgf("接收到事件:  %v", utils.BytesToString(payload))
		handler(payload, wssc)
	}
}

func (wssc *WSSCaller) nextSeq() uint64 {
	return atomic.AddUint64(&wssc.seq, 1)
}

// CallApi 发送ws请求
func (wssc *WSSCaller) CallApi(req zero.APIRequest) (zero.APIResponse, error) {
	ch := make(chan zero.APIResponse, 1)
	req.Echo = wssc.nextSeq()
	wssc.seqMap.Store(req.Echo, ch)

	data, _ := json.Marshal(&req)
	// send message
	wssc.mu.Lock() // websocket write is not goroutine safe
	err := wssc.conn.WriteMessage(websocket.BinaryMessage, data)
	wssc.mu.Unlock()
	if err != nil {
		log.Warn().Str("name", "wss").Err(err).Msg("向WebsocketServer发送API请求失败")
		return nullResponse, err
	}
	log.Debug().Str("name", "wss").Msgf("向服务器发送请求: %v", utils.BytesToString(data))

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
