package zero

import (
	"fmt"
	"reflect"
	"sync"

	"ZeroBot/message"

	"github.com/sirupsen/logrus"
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
	once    sync.Once
	message string
	Err     error
}

// func (ctx )
func (ctx *Ctx) GetMatcherMetadata() MatcherMetadata {
	return MatcherMetadata{
		PluginName: ctx.ma.Engine.MetaData.Name,
	}
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
			ctx.Event.GroupID == ctx2.Event.GroupID // 私聊时GroupID为0，也相等
	}
}

// Send 快捷发送消息/合并转发
func (ctx *Ctx) Send(msg interface{}) message.MessageID {
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
		if event.GroupID != 0 {
			return message.NewMessageIDFromInteger(ctx.SendGroupForwardMessage(event.GroupID, m).Get("message_id").Int())
		}
		return message.NewMessageIDFromInteger(ctx.SendPrivateForwardMessage(event.UserID, m).Get("message_id").Int())
	}
	if event.DetailType == "guild" {
		return message.NewMessageIDFromString(ctx.SendGuildChannelMessage(event.GuildID, event.ChannelID, msg))
	}
	if event.GroupID != 0 {
		return message.NewMessageIDFromInteger(ctx.SendGroupMessage(event.GroupID, msg))
	}
	return message.NewMessageIDFromInteger(ctx.SendPrivateMessage(event.UserID, msg))
}

// SendChain 快捷发送消息/合并转发-消息链
func (ctx *Ctx) SendChain(msg ...message.MessageSegment) message.MessageID {
	return ctx.Send((message.Message)(msg))
}

func (ctx *Ctx) SendError(text string, err error) message.MessageID {
	ctx.Err = err
	name := ctx.ma.Engine.MetaData.Name
	if name == "" {
		name = "default"
	}
	msg := fmt.Sprintf("[%s] %s err: %v", name, text, err)
	logrus.Errorln(msg)

	return ctx.SendChain(message.At(ctx.Event.UserID), message.Text(msg))
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
	return ctx.Event.Message.ExtractPlainText()
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
		}
	})
	return ctx.message
}
