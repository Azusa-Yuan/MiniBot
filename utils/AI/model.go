package ai

import "sync"

type IntroduceManger struct {
	IntroduceMap map[string]string `json:"introduce_map"`
	sync.RWMutex
}

type Config struct {
	Key string `yaml:"key"`
}

var BidMap map[int64]string
