package monitor

import (
	"MiniBot/utils/path"
	"fmt"
)

type config struct {
	Pprof bool `yaml:"pprof"`
}

func init() {
	fmt.Println(path.GetPluginDataPath())
}
