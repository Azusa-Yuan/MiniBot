package pcrjjc3

import (
	"sync"
)

// 配置结构体
type Config struct {
	GlobalPush     bool   `yaml:"global_push"`
	ScheduleThread int    `yaml:"schedule_thread"`
	ScheduleTime   int    `yaml:"schedule_time"`
	AccountLimit   int    `yaml:"account_limit"`
	Proxy          string `yaml:"proxy"`
	sync.RWMutex
}

func (conf *Config) setGloablPush(b bool) {
	conf.Lock()
	defer conf.Unlock()
	conf.GlobalPush = b
}

var modeLoc = map[string]int{
	"arena_on":       1,
	"grand_arena_on": 2,
	"rank_up":        3,
	"login_remind":   4,
}
