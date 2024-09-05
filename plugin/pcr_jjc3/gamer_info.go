package pcrjjc3

import (
	"fmt"
	"sync"
)

type GamerInfoUnit struct {
	ArenaRank      int64
	GrandArenaRank int64
	GamerName      string
	TimeStamp      int64
}

type GamerInfoManage struct {
	GamerInfoMap    map[string]*GamerInfoUnit
	TmpGamerInfoMap map[string]*GamerInfoUnit
	sync.RWMutex
}

func NewGamerInfoManage() GamerInfoManage {
	return GamerInfoManage{GamerInfoMap: map[string]*GamerInfoUnit{},
		TmpGamerInfoMap: map[string]*GamerInfoUnit{},
	}
}

func (g *GamerInfoManage) getSnapshot(idList []string) string {
	g.RLock()
	defer g.RUnlock()
	msg := ""
	for i, id := range idList {
		if info, ok := g.GamerInfoMap[id]; ok {
			msg += fmt.Sprintf("\r\n%d %s %s jjc:%d pjjc:%d", i+1, id, info.GamerName, info.ArenaRank, info.GrandArenaRank)
		}
	}
	msg += fmt.Sprintf("\r\n该排名有延时(最大为%ds)，仅供参考", config.ScheduleTime*3)
	return msg
}
