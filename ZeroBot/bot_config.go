package zero

import "time"

// Config is config of zero bot
type Config struct {
	NickName       []string           `json:"nickname"  yaml:"nickname" mapstructure:"nickname"`                        // 机器人名称
	CommandPrefix  string             `json:"command_prefix" yaml:"command_prefix"  mapstructure:"command_prefix"`      // 触发命令
	SuperUsers     []int64            `json:"super_users" yaml:"super_users" mapstructure:"super_users"`                // 超级用户
	RingLen        uint               `json:"ring_len" yaml:"ring_len" mapstructure:"ring_len"`                         // 事件环长度 (默认关闭)
	Latency        time.Duration      `json:"latency" yaml:"latency" mapstructure:"latency"`                            // 事件处理延迟 (延迟 latency 再处理事件，在 ring 模式下不可低于 1ms)
	MaxProcessTime time.Duration      `json:"max_process_time" yaml:"max_process_time" mapstructure:"max_process_time"` // 事件最大处理时间 (默认4min)
	MarkMessage    bool               `json:"mark_message" yaml:"mark_message" mapstructure:"mark_message"`             // 自动标记消息为已读
	Driver         []Driver           `json:"-"  yaml:"-"`                                                              // 通信驱动
	InstanceMap    map[int64]Instance `json:"instance" yaml:"instance" mapstructure:"instance"`
}

// BotConfig 运行中bot的配置
var BotConfig Config

type Instance struct {
	SelfID        int64    `json:"seld_id" mapstructure:"seld_id"`                                      // 机器人QQ号
	NickName      []string `json:"nickname"  yaml:"nickname" mapstructure:"nickname"`                   // 机器人名称
	CommandPrefix string   `json:"command_prefix" yaml:"command_prefix"  mapstructure:"command_prefix"` // 触发命令
	SuperUsers    []int64  `json:"super_users" yaml:"super_users" mapstructure:"super_users"`           // 超级用户
}

func (c Config) GetNickName(selfID int64) []string {
	if instance, ok := c.InstanceMap[selfID]; ok {
		return instance.NickName
	}
	return c.NickName
}

func (c Config) GetSuperUser(selfID int64) []int64 {
	if instance, ok := c.InstanceMap[selfID]; ok {
		return instance.SuperUsers
	}
	return c.SuperUsers
}
