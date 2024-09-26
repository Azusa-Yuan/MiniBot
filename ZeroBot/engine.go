package zero

import "github.com/rs/zerolog/log"

// New 生成空引擎  只允许默认引擎使用
func unew() *Engine {
	return &Engine{
		PreHandler:  []Rule{},
		MidHandler:  []Rule{},
		PostHandler: []Handler{},
	}
}

type MetaData struct {
	Name  string
	Help  string
	Level uint
}

func NewTemplate(metaData *MetaData) (e *Engine) {
	if metaData.Name == "" {
		log.Fatal().Msg("禁止注册没有命名的插件")
	}
	e = &Engine{
		MetaData:    metaData,
		PreHandler:  []Rule{},
		MidHandler:  []Rule{},
		PostHandler: []Handler{},
	}
	e.UsePreHandler(
		func(ctx *Ctx) bool {
			// 防止自触发
			return ctx.Event.UserID != ctx.Event.SelfID || ctx.Event.PostType != "message"
		},
	)
	return
}

var defaultEngine = unew()

// Engine is the pre_handler, post_handler manager
type Engine struct {
	MetaData    *MetaData
	PreHandler  []Rule
	MidHandler  []Rule
	PostHandler []Handler
	Block       bool
	Matchers    []*Matcher
}

// Delete 移除该 Engine 注册的所有 Matchers
func (e *Engine) Delete() {
	for _, m := range e.Matchers {
		m.Delete()
	}
}

func (e *Engine) SetBlock(Block bool) *Engine {
	e.Block = Block
	return e
}

func (e *Engine) GetMetaDate(Block bool) *MetaData {
	return e.MetaData
}

func (e *Engine) SetMetaDate(MetaData *MetaData) {
	e.MetaData = MetaData
}

// UsePreHandler 向该 Engine 添加新 PreHandler(Rule),
// 会在 Rule 判断前触发，如果 PreHandler
// 没有通过，则 Rule, Matcher 不会触发
//
// 可用于分群组管理插件等
func (e *Engine) UsePreHandler(rules ...Rule) {
	e.PreHandler = append(e.PreHandler, rules...)
}

// UseMidHandler 向该 Engine 添加新 MidHandler(Rule),
// 会在 Rule 判断后， Matcher 触发前触发，如果 MidHandler
// 没有通过，则 Matcher 不会触发
//
// 可用于速率限制等
func (e *Engine) UseMidHandler(rules ...Rule) {
	e.MidHandler = append(e.MidHandler, rules...)
}

// UsePostHandler 向该 Engine 添加新 PostHandler(Rule),
// 会在 Matcher 触发后触发，如果 PostHandler 返回 false,
// 则后续的 post handler 不会触发
//
// 可用于反并发等
func (e *Engine) UsePostHandler(handler ...Handler) {
	e.PostHandler = append(e.PostHandler, handler...)
}

// On 添加新的指定消息类型的匹配器(默认Engine)
func On(typ string, rules ...Rule) *Matcher { return defaultEngine.On(typ, rules...) }

// On 添加新的指定消息类型的匹配器 格式post_type/detail_type/sub_type
func (e *Engine) On(typ string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type(typ),
		Rules:  rules,
		Engine: e,
		Mark:   typ,
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnMessage 消息触发器
func OnMessage(rules ...Rule) *Matcher { return On("message", rules...) }

// OnMessage 消息触发器
func (e *Engine) OnMessage(rules ...Rule) *Matcher { return e.On("message", rules...) }

// OnNotice 系统提示触发器
func OnNotice(rules ...Rule) *Matcher { return On("notice", rules...) }

// OnNotice 系统提示触发器
func (e *Engine) OnNotice(rules ...Rule) *Matcher { return e.On("notice", rules...) }

// OnRequest 请求消息触发器
func OnRequest(rules ...Rule) *Matcher { return On("request", rules...) }

// OnRequest 请求消息触发器
func (e *Engine) OnRequest(rules ...Rule) *Matcher { return On("request", rules...) }

// OnMetaEvent 元事件触发器
func OnMetaEvent(rules ...Rule) *Matcher { return On("meta_event", rules...) }

// OnMetaEvent 元事件触发器
func (e *Engine) OnMetaEvent(rules ...Rule) *Matcher { return On("meta_event", rules...) }

// OnPrefix 前缀触发器  ctx.State["args"]   ctx.State["prefix"]
func OnPrefix(prefix string, rules ...Rule) *Matcher { return defaultEngine.OnPrefix(prefix, rules...) }

// OnPrefix 前缀触发器  ctx.State["args"]   ctx.State["prefix"]
func (e *Engine) OnPrefix(prefix string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{PrefixRule(prefix)}, rules...),
		Engine: e,
		Mark:   prefix,
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnSuffix 后缀触发器
func OnSuffix(suffix string, rules ...Rule) *Matcher { return defaultEngine.OnSuffix(suffix, rules...) }

// OnSuffix 后缀触发器
func (e *Engine) OnSuffix(suffix string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{SuffixRule(suffix)}, rules...),
		Engine: e,
		Mark:   suffix,
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnCommand 命令触发器
func OnCommand(commands string, rules ...Rule) *Matcher {
	return defaultEngine.OnCommand(commands, rules...)
}

// OnCommand 命令触发器
func (e *Engine) OnCommand(commands string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{CommandRule(commands)}, rules...),
		Engine: e,
		Mark:   commands,
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnRegex 正则触发器
func OnRegex(regexPattern string, rules ...Rule) *Matcher {
	return OnMessage(append([]Rule{RegexRule(regexPattern)}, rules...)...)
}

// OnRegex 正则触发器
func (e *Engine) OnRegex(regexPattern string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{RegexRule(regexPattern)}, rules...),
		Engine: e,
		Mark:   regexPattern,
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnKeyword 关键词触发器
func OnKeyword(keyword string, rules ...Rule) *Matcher {
	return defaultEngine.OnKeyword(keyword, rules...)
}

// OnKeyword 关键词触发器
func (e *Engine) OnKeyword(keyword string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{KeywordRule(keyword)}, rules...),
		Engine: e,
		Mark:   keyword,
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnFullMatch 完全匹配触发器
func OnFullMatch(src string, rules ...Rule) *Matcher {
	return defaultEngine.OnFullMatch(src, rules...)
}

// OnFullMatch 完全匹配触发器
func (e *Engine) OnFullMatch(src string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{FullMatchRule(src)}, rules...),
		Engine: e,
		Mark:   src,
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnFullMatchGroup 完全匹配触发器组
func OnFullMatchGroup(src []string, rules ...Rule) *Matcher {
	return defaultEngine.OnFullMatchGroup(src, rules...)
}

// OnFullMatchGroup 完全匹配触发器组
func (e *Engine) OnFullMatchGroup(src []string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{FullMatchRule(src...)}, rules...),
		Engine: e,
		Mark:   src[0],
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnKeywordGroup 关键词触发器组
func OnKeywordGroup(keywords []string, rules ...Rule) *Matcher {
	return defaultEngine.OnKeywordGroup(keywords, rules...)
}

// OnKeywordGroup 关键词触发器组
func (e *Engine) OnKeywordGroup(keywords []string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{KeywordRule(keywords...)}, rules...),
		Engine: e,
		Mark:   keywords[0],
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnCommandGroup 命令触发器组
func OnCommandGroup(commands []string, rules ...Rule) *Matcher {
	return defaultEngine.OnCommandGroup(commands, rules...)
}

// OnCommandGroup 命令触发器组
func (e *Engine) OnCommandGroup(commands []string, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{CommandRule(commands...)}, rules...)...)
}

// OnPrefixGroup 前缀触发器组  ctx.State["args"]   ctx.State["prefix"]
func OnPrefixGroup(prefix []string, rules ...Rule) *Matcher {
	return defaultEngine.OnPrefixGroup(prefix, rules...)
}

// OnPrefixGroup 前缀触发器组 ctx.State["args"]   ctx.State["prefix"]
func (e *Engine) OnPrefixGroup(prefix []string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{PrefixRule(prefix...)}, rules...),
		Engine: e,
		Mark:   prefix[0],
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnSuffixGroup 后缀触发器组
func OnSuffixGroup(suffix []string, rules ...Rule) *Matcher {
	return defaultEngine.OnSuffixGroup(suffix, rules...)
}

// OnSuffixGroup 后缀触发器组
func (e *Engine) OnSuffixGroup(suffix []string, rules ...Rule) *Matcher {
	matcher := &Matcher{
		Type:   Type("message"),
		Rules:  append([]Rule{SuffixRule(suffix...)}, rules...),
		Engine: e,
		Mark:   suffix[0],
	}
	e.Matchers = append(e.Matchers, matcher)
	return StoreMatcher(matcher)
}

// OnShell shell命令触发器
func OnShell(command string, model interface{}, rules ...Rule) *Matcher {
	return defaultEngine.OnShell(command, model, rules...)
}

// OnShell shell命令触发器
func (e *Engine) OnShell(command string, model interface{}, rules ...Rule) *Matcher {
	return e.On("message", append([]Rule{ShellRule(command, model)}, rules...)...)
}

func init() {
	defaultEngine.UsePreHandler(
		func(ctx *Ctx) bool {
			// 防止自触发
			return ctx.Event.UserID != ctx.Event.SelfID || ctx.Event.PostType != "message"
		},
	)
}
