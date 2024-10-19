package zero

import (
	"fmt"
	"reflect"
	"sync"

	"ZeroBot/message"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

// Ctx represents the Context which hold the event.
// 代表上下文  ctx包含了ws通信的信息 todo  实现完整的上下文
type Ctx struct {
	ma     *Matcher
	Event  *Event
	State  State
	caller APICaller
	stop   bool // 用于终止会话

	// lazy message
	once       sync.Once
	message    string
	rawMessage string
	Err        error
}

func (ctx *Ctx) GetMatcherMetadata() MatcherMetadata {
	meta := MatcherMetadata{
		MatcherName: ctx.ma.Mark,
	}
	if ctx.ma.Engine.Metadata == nil {
		meta.PluginName = "default"
	} else {
		meta.PluginName = ctx.ma.Engine.Metadata.Name
	}
	return meta
}

// GetMatcher ...
func (ctx *Ctx) GetMatcher() *Matcher {
	return ctx.ma
}

// decoder 反射获取的数据
type decoder []dec

type dec struct {
	index int
	key   string
}

// decoder 缓存
var decoderCache = sync.Map{}

// Parse 将 Ctx.State 映射到结构体
func (ctx *Ctx) Parse(model interface{}) (err error) {
	var (
		rv       = reflect.ValueOf(model).Elem()
		t        = rv.Type()
		modelDec decoder
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parse state error: %v", r)
		}
	}()
	d, ok := decoderCache.Load(t)
	if ok {
		modelDec = d.(decoder)
	} else {
		modelDec = decoder{}
		for i := 0; i < t.NumField(); i++ {
			t1 := t.Field(i)
			if key, ok := t1.Tag.Lookup("zero"); ok {
				modelDec = append(modelDec, dec{
					index: i,
					key:   key,
				})
			}
		}
		decoderCache.Store(t, modelDec)
	}
	for _, d := range modelDec { // decoder类型非小内存，无法被编译器优化为快速拷贝
		rv.Field(d.index).Set(reflect.ValueOf(ctx.State[d.key]))
	}
	return nil
}

// CheckSession 判断会话连续性
func (ctx *Ctx) CheckSession() Rule {
	return func(ctx2 *Ctx) bool {
		return ctx.Event.UserID == ctx2.Event.UserID &&
			ctx.Event.GroupID == ctx2.Event.GroupID &&
			ctx.Event.SelfID == ctx2.Event.SelfID // 私聊时GroupID为0，也相等
	}
}

// Send 快捷发送消息/合并转发
func (ctx *Ctx) Send(msg interface{}) (message.MessageID, error) {
	event := ctx.Event
	m, ok := msg.(message.Message)
	if !ok {
		var p *message.Message
		p, ok = msg.(*message.Message)
		if ok {
			m = *p
		}
	}
	if ok && len(m) > 0 && m[0].Type == "node" && event.DetailType != "guild" {
		var resp gjson.Result
		var err error
		if event.GroupID != 0 {
			resp, err = ctx.SendGroupForwardMessage(event.GroupID, m)

		} else {
			resp, err = ctx.SendPrivateForwardMessage(event.UserID, m)
		}

		if err != nil {
			return message.MessageID{}, err
		}
		return message.NewMessageIDFromInteger(resp.Get("message_id").Int()), nil
	}
	if event.DetailType == "guild" {
		resp, err := ctx.SendGuildChannelMessage(event.GuildID, event.ChannelID, msg)
		return message.NewMessageIDFromString(resp), err
	}
	var resp int64
	var err error
	if event.GroupID != 0 {
		resp, err = ctx.SendGroupMessage(event.GroupID, msg)
	} else {
		resp, err = ctx.SendPrivateMessage(event.UserID, msg)
	}

	if err != nil {
		return message.MessageID{}, err
	}
	return message.NewMessageIDFromInteger(resp), nil
}

// SendChain 快捷发送消息/合并转发-消息链
func (ctx *Ctx) SendChain(msg ...message.MessageSegment) (message.MessageID, error) {
	return ctx.Send((message.Message)(msg))
}

func (ctx *Ctx) SendError(err error, msgs ...message.MessageSegment) (message.MessageID, error) {
	ctx.Err = err
	metadata := ctx.GetMatcherMetadata()
	log.Error().CallerSkipFrame(1).Str("name", metadata.PluginName).Str("controller", metadata.MatcherName).Err(err).Msg("")
	msgs = append(message.Message{message.At(ctx.Event.UserID), message.Text(err)}, msgs...)
	return ctx.SendChain(msgs...)
}

// Echo 向自身分发虚拟事件
func (ctx *Ctx) Echo(response []byte) {
	if BotConfig.RingLen != 0 {
		evchan.processEvent(response, ctx.caller)
	} else {
		processEventAsync(response, ctx.caller, BotConfig.MaxProcessTime)
	}
}

// FutureEvent ...
func (ctx *Ctx) FutureEvent(Type string, rule ...Rule) *FutureEvent {
	return ctx.ma.FutureEvent(Type, rule...)
}

// Get ..
func (ctx *Ctx) Get(prompt string) string {
	if prompt != "" {
		ctx.Send(prompt)
	}
	return (<-ctx.FutureEvent("message", ctx.CheckSession()).Next()).Event.RawMessage
}

// ExtractPlainText 提取消息中的纯文本
func (ctx *Ctx) ExtractPlainText() string {
	if ctx == nil || ctx.Event == nil || ctx.Event.Message == nil {
		return ""
	}
	ctx.once.Do(func() {
		if ctx.Event != nil && ctx.Event.Message != nil {
			ctx.message = ctx.Event.Message.String()
			ctx.rawMessage = ctx.Event.Message.ExtractPlainText()
		}
	})
	return ctx.rawMessage
}

// Block 匹配成功后阻止后续触发
func (ctx *Ctx) Block() {
	ctx.ma.SetBlock(true)
}

// Stop 中止会话
func (ctx *Ctx) Stop() {
	ctx.stop = true
}

// NoTimeout 处理时不设超时
func (ctx *Ctx) NoTimeout() {
	ctx.ma.NoTimeout = true
}

// MessageString 字符串消息便于Regex
func (ctx *Ctx) MessageString() string {
	ctx.once.Do(func() {
		if ctx.Event != nil && ctx.Event.Message != nil {
			ctx.message = ctx.Event.Message.String()
			ctx.rawMessage = ctx.Event.Message.ExtractPlainText()
		}
	})
	return ctx.message
}

// 待完善
func (ctx *Ctx) ReceptionToSend() message.Message {
	if ctx.Event == nil || ctx.Event.Message == nil {
		return nil
	}
	tmpMessage := ctx.Event.Message
	for i, segment := range tmpMessage {
		switch segment.Type {
		case "image":
			tmpMessage[i] = message.Image(segment.Data["url"])
		case "text":
			tmpMessage[i] = message.Text(segment.Data["text"])
		}
	}
	return tmpMessage
}
