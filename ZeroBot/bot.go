package zero

import (
	"encoding/json"
	"fmt"
	"hash/crc64"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FloatTech/ttl"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	"ZeroBot/message"
	"ZeroBot/utils"
)

// APICallers 所有的APICaller列表， 通过self-ID映射
var APICallers callerMap

// APICaller is the interface of CallApi
type APICaller interface {
	CallApi(request APIRequest) (APIResponse, error)
}

// Driver 与OneBot通信的驱动，使用driver.DefaultWebSocketDriver
type Driver interface {
	Connect()
	Listen(func([]byte, APICaller))
}

var (
	evchan    eventChan // evring 事件环
	isrunning uintptr
)

func runinit(op *Config) {
	if op.MaxProcessTime == 0 {
		op.MaxProcessTime = time.Minute * 4
	}
	BotConfig = *op
	if op.RingLen == 0 {
		return
	}
	evchan = newchan(op.RingLen)
	evchan.loop(op.Latency, op.MaxProcessTime, processEventAsync)
}

func (op *Config) directlink(b []byte, c APICaller) {
	go func() {
		if op.Latency != 0 {
			time.Sleep(op.Latency)
		}
		processEventAsync(b, c, op.MaxProcessTime)
	}()
}

// Run 主函数初始化
func Run(op *Config) {
	if !atomic.CompareAndSwapUintptr(&isrunning, 0, 1) {
		log.Warn().Str("name", "bot").Msg("已忽略重复调用的 Run")
	}
	runinit(op)
	// listenCallback 有两种不同的处理上报event策略
	// 第一种是收到onebot消息直接处理
	// 另一种是收到消息放到事件管道里处理
	listenCallback := op.directlink
	if op.RingLen != 0 {
		listenCallback = evchan.processEvent
	}
	for _, driver := range op.Driver {
		driver.Connect()
		go driver.Listen(listenCallback)
	}
}

// RunAndBlock 主函数初始化并阻塞
//
// 除最后一个Driver都用go实现非阻塞
func RunAndBlock(op *Config, preblock func()) {
	if !atomic.CompareAndSwapUintptr(&isrunning, 0, 1) {
		log.Warn().Str("name", "bot").Msg("已忽略重复调用的 RunAndBlock")
	}
	// 初始化消息处理机制
	runinit(op)
	listenCallback := op.directlink
	if op.RingLen != 0 {
		listenCallback = evchan.processEvent
	}
	switch len(op.Driver) {
	case 0:
		return
	case 1:
		op.Driver[0].Connect()
		if preblock != nil {
			preblock()
		}
		op.Driver[0].Listen(listenCallback)
	default:
		i := 0
		for ; i < len(op.Driver)-1; i++ {
			op.Driver[i].Connect()
			go op.Driver[i].Listen(listenCallback)
		}
		op.Driver[i].Connect()
		if preblock != nil {
			preblock()
		}
		// listen是for死循环  会阻塞
		op.Driver[i].Listen(listenCallback)
	}
}

var (
	triggeredMessages   = ttl.NewCache[int64, []message.MessageID](time.Minute * 5)
	triggeredMessagesMu = sync.Mutex{}
)

type messageLogger struct {
	msgid  message.MessageID
	caller APICaller
}

// CallApi 记录被触发的回复消息
func (m *messageLogger) CallApi(request APIRequest) (rsp APIResponse, err error) {
	rsp, err = m.caller.CallApi(request)
	if err != nil {
		return
	}
	id := rsp.Data.Get("message_id")
	if id.Exists() {
		mid := m.msgid.ID()
		triggeredMessagesMu.Lock()
		defer triggeredMessagesMu.Unlock()
		triggeredMessages.Set(mid,
			append(
				triggeredMessages.Get(mid),
				message.NewMessageIDFromString(id.String()),
			),
		)
	}
	return
}

// GetTriggeredMessages 获取被 id 消息触发的回复消息 id
func GetTriggeredMessages(id message.MessageID) []message.MessageID {
	triggeredMessagesMu.Lock()
	defer triggeredMessagesMu.Unlock()
	return triggeredMessages.Get(id.ID())
}

// processEventAsync 处理事件, 异步调用匹配 mather
func processEventAsync(response []byte, caller APICaller, maxwait time.Duration) {
	var event Event
	_ = json.Unmarshal(response, &event)
	event.RawEvent = gjson.Parse(utils.BytesToString(response))
	var msgid message.MessageID
	messageID, err := strconv.ParseInt(utils.BytesToString(event.RawMessageID), 10, 64)
	if err == nil {
		event.MessageID = messageID
		msgid = message.NewMessageIDFromInteger(messageID)
	} else if event.MessageType == "guild" {
		// 是 guild 消息，进行如下转换以适配非 guild 插件
		// MessageID 填为 string
		event.MessageID, _ = strconv.Unquote(utils.BytesToString(event.RawMessageID))
		// 伪造 GroupID
		crc := crc64.New(crc64.MakeTable(crc64.ISO))
		crc.Write(utils.StringToBytes(event.GuildID))
		crc.Write(utils.StringToBytes(event.ChannelID))
		r := int64(crc.Sum64() & 0x7fff_ffff_ffff_ffff) // 确保为正数
		if r <= 0xffff_ffff {
			r |= 0x1_0000_0000 // 确保不与正常号码重叠
		}
		event.GroupID = r
		// 伪造 UserID
		crc.Reset()
		crc.Write(utils.StringToBytes(event.TinyID))
		r = int64(crc.Sum64() & 0x7fff_ffff_ffff_ffff) // 确保为正数
		if r <= 0xffff_ffff {
			r |= 0x1_0000_0000 // 确保不与正常号码重叠
		}
		event.UserID = r
		if event.Sender != nil {
			event.Sender.ID = r
		}
		msgid = message.NewMessageIDFromString(event.MessageID.(string))
	}

	switch event.PostType { // process DetailType
	case "message", "message_sent":
		event.DetailType = event.MessageType
	case "notice":
		event.DetailType = event.NoticeType
		preprocessNoticeEvent(&event)
		printNoticeLog(&event)
	case "request":
		event.DetailType = event.RequestType
		printRequestLog(&event)
	}
	if event.PostType == "message" {
		preprocessMessageEvent(&event)
		printMessageLog(&event)
	}
	ctx := &Ctx{
		Event:  &event,
		State:  State{},
		caller: &messageLogger{msgid: msgid, caller: caller},
	}
	matcherLock.Lock()
	defer matcherLock.Unlock()
	if hasMatcherListChanged {
		matcherListForRanging = make([]*Matcher, len(matcherList))
		copy(matcherListForRanging, matcherList)
		hasMatcherListChanged = false
	}

	// 全局开始前中间件
	GolbaleMiddleware.HandleBegin(ctx)
	if ctx.stop {
		return
	}
	go match(ctx, matcherListForRanging, maxwait)
}

// match 匹配规则，处理事件
func match(ctx *Ctx, matchers []*Matcher, maxwait time.Duration) {
	if BotConfig.MarkMessage && ctx.Event.MessageID != nil {
		ctx.MarkThisMessageAsRead()
	}
	gorule := func(rule Rule) <-chan bool {
		ch := make(chan bool, 1)
		go func() {
			defer func() {
				close(ch)
				if pa := recover(); pa != nil {
					log.Error().Str("name", "bot").Msgf("execute rule err: %v\n%v", pa, utils.BytesToString(debug.Stack()))
				}
			}()
			ch <- rule(ctx)
		}()
		return ch
	}
	gohandler := func(h Handler) <-chan struct{} {
		ch := make(chan struct{}, 1)
		go func() {
			defer func() {
				close(ch)
				if pa := recover(); pa != nil {
					log.Error().Str("name", "bot").Msgf("execute handler err: %v\n%v", pa, utils.BytesToString(debug.Stack()))
				}
			}()
			h(ctx)
			ch <- struct{}{}
		}()
		return ch
	}
	waitTimer := time.NewTimer(maxwait)
	defer waitTimer.Stop()
loop:
	for _, matcher := range matchers {
		if !matcher.Type(ctx) {
			continue
		}
		for k := range ctx.State { // Clear State
			delete(ctx.State, k)
		}
		// 这里对matcher进行了深拷贝
		m := matcher.copy()
		ctx.ma = m

		// pre handler
		// 因为matcher 匹配在matcher的rule中，所以engine.PreHandler无论如何都会执行
		if m.Engine != nil {
			for _, handler := range m.Engine.PreHandler {
				c := gorule(handler)
				for {
					select {
					case ok := <-c:
						if ctx.stop {
							break loop
						}
						if !ok { // 有 pre handler 未满足
							continue loop
						}
					case <-waitTimer.C:
						if m.NoTimeout { // 不设超时限制
							waitTimer.Reset(maxwait)
							continue
						}
						log.Warn().Str("name", "bot").Msg("PreHandler 处理达到最大时延, 退出")
						break loop
					}
					break
				}
			}
		}

		for _, rule := range m.Rules {
			c := gorule(rule)
			for {
				select {
				case ok := <-c:
					if ctx.stop {
						break loop
					}
					if !ok { // 有 Rule 的条件未满足
						continue loop
					}
				case <-waitTimer.C:
					if m.NoTimeout { // 不设超时限制
						waitTimer.Reset(maxwait)
						continue
					}
					log.Warn().Str("name", "bot").Msg("rule 处理达到最大时延, 退出")
					break loop
				}
				break
			}
		}

		// mid handler
		if m.Engine != nil {
			for _, handler := range m.Engine.MidHandler {
				c := gorule(handler)
				for {
					select {
					case ok := <-c:
						if ctx.stop {
							break loop
						}
						if !ok { // 有 mid handler 未满足
							continue loop
						}
					case <-waitTimer.C:
						if m.NoTimeout { // 不设超时限制
							waitTimer.Reset(maxwait)
							continue
						}
						log.Warn().Str("name", "bot").Msg("midHandler 处理达到最大时延, 退出")
						break loop
					}
					break
				}
			}
		}

		// 全局before hander处理
		GolbaleMiddleware.HandleBefore(ctx)

		if m.Handler != nil {
			c := gohandler(m.Handler)
			for {
				select {
				case <-c: // 处理事件
				case <-waitTimer.C:
					if m.NoTimeout { // 不设超时限制
						waitTimer.Reset(maxwait)
						continue
					}
					log.Warn().Str("name", "bot").Msg("Handler 处理达到最大时延, 退出")
					break loop
				}
				break
			}
		}
		if matcher.Temp { // 临时 Matcher 删除
			matcher.Delete()
		}

		// 全局after hander处理
		GolbaleMiddleware.HandleAfter(ctx)

		if m.Engine != nil {
			// post handler
			for _, handler := range m.Engine.PostHandler {
				c := gohandler(handler)
				for {
					select {
					case <-c:
					case <-waitTimer.C:
						if m.NoTimeout { // 不设超时限制
							waitTimer.Reset(maxwait)
							continue
						}
						log.Warn().Str("name", "bot").Msg("postHandler 处理达到最大时延, 退出")
						break loop
					}
					break
				}
			}
		}

		if ctx.stop {
			break loop
		}

		if m.Block { // 阻断后续
			break loop
		}
	}
	// 全局结束后中间件
	GolbaleMiddleware.HandleEnd(ctx)
}

// preprocessMessageEvent 返回信息事件
func preprocessMessageEvent(e *Event) {
	e.Message = message.ParseMessage(e.NativeMessage)

	processAt := func() { // 处理是否at机器人
		e.IsToMe = false
		for i, m := range e.Message {
			if m.Type == "at" {
				qq, _ := strconv.ParseInt(m.Data["qq"], 10, 64)
				if qq == e.SelfID {
					e.IsToMe = true
					e.Message = append(e.Message[:i], e.Message[i+1:]...)
					return
				}
			}
		}
		if len(e.Message) == 0 || e.Message[0].Type != "text" {
			return
		}
		first := e.Message[0]
		first.Data["text"] = strings.TrimLeft(first.Data["text"], " ") // Trim!
		text := first.Data["text"]
		for _, nickname := range BotConfig.GetNickName(e.SelfID) {
			if strings.HasPrefix(text, nickname) {
				e.IsToMe = true
				first.Data["text"] = text[len(nickname):]
				return
			}
		}
	}

	switch {
	case e.DetailType == "group":
		processAt()
	case e.DetailType == "guild" && e.SubType == "channel":
		processAt()
	default:
		e.IsToMe = true // 私聊也判断为at
	}
	if len(e.Message) > 0 && e.Message[0].Type == "text" { // Trim Again!
		e.Message[0].Data["text"] = strings.TrimLeft(e.Message[0].Data["text"], " ")
	}
}

// preprocessNoticeEvent 更新事件
func preprocessNoticeEvent(e *Event) {
	if e.SubType == "poke" || e.SubType == "lucky_king" {
		e.IsToMe = e.TargetID == e.SelfID
	} else {
		e.IsToMe = e.UserID == e.SelfID
	}
}

// GetBot 获取指定的bot (Ctx)实例
func GetBot(id int64) (*Ctx, error) {
	caller, ok := APICallers.Load(id)
	if !ok {
		return nil, fmt.Errorf("[bot] %d bot not found", id)
	}
	return &Ctx{caller: caller}, nil
}

// RangeBot 遍历所有bot (Ctx)实例
//
// 单次操作返回 true 则继续遍历，否则退出
func RangeBot(iter func(id int64, ctx *Ctx) bool) {
	APICallers.Range(func(key int64, value APICaller) bool {
		return iter(key, &Ctx{caller: value})
	})
}

// GetFirstSuperUser 在 qqs 中获得 SuperUsers 列表的首个 qq
//
// 找不到返回 -1
func (c *Config) GetFirstSuperUser(qqs ...int64) int64 {
	m := make(map[int64]struct{}, len(qqs)*4)
	for _, qq := range qqs {
		m[qq] = struct{}{}
	}
	for _, qq := range c.SuperUsers {
		if _, ok := m[qq]; ok {
			return qq
		}
	}
	return -1
}
