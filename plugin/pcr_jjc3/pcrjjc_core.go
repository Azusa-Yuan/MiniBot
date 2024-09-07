package pcrjjc3

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

var (
	gamerInfoManage = NewGamerInfoManage()
	queryLock       sync.Mutex
)

func checkServer(cx string) (client *pcrclient, err error) {
	var ok bool
	if client, ok = clientMap[cx]; !ok {
		err = fmt.Errorf("不支持该服务器或服务器选择错误")
	}
	return
}

func userQuery(id string) (res string, err error) {
	resp, err := query(id, false)
	if err != nil {
		return
	}
	userInfo := resp.Get("user_info")
	deepDomain := resp.Get("quest_info").Get("talent_quest")

	timeStamp := userInfo.Get("last_login_time").Int()
	timeStr := time.Unix(timeStamp, 0).Format("2006-01-02 15:04:05")

	knight_exp := userInfo.Get("princess_knight_rank_total_exp").Int()

	res = fmt.Sprintf(`
名字: %s
区服：%s
jjc排名: %v
pjjc排名: %v
最后登录: %v
竞技场场次: %v
公主竞技场场次: %v
公主骑士Rank: %v
火: %v   水: %v
风: %v    光: %v
暗: %v `, userInfo.Get("user_name").Str, cxMap[id[:1]], userInfo.Get("arena_rank"), userInfo.Get("grand_arena_rank"), timeStr,
		userInfo.Get("arena_group"), userInfo.Get("grand_arena_group"), calKnightRank(int(knight_exp)),
		calculateDomain(int(deepDomain.Get("0").Get("clear_count").Int())), calculateDomain(int(deepDomain.Get("1").Get("clear_count").Int())),
		calculateDomain(int(deepDomain.Get("2").Get("clear_count").Int())), calculateDomain(int(deepDomain.Get("3").Get("clear_count").Int())),
		calculateDomain(int(deepDomain.Get("4").Get("clear_count").Int())))
	return
}

func query(id string, save bool) (gjson.Result, error) {
	client, err := checkServer(id[:1])
	if err != nil {
		return gjson.Result{}, err
	}

	request := map[string]interface{}{}
	account, _ := strconv.Atoi(id)
	request["target_viewer_id"] = account

	// 双重校验锁 极大阻止发生错误时继续发送请求
	if !client.login {
		// time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		queryLock.Lock()
		if !client.login {
			client.Login()
		}
		queryLock.Unlock()
	}
	resp := gjson.Result{}
	resp, err = client.CallApi("/profile/get_profile", request)
	if err != nil {
		// 错误  重新执行一次
		if !client.login {
			// time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
			queryLock.Lock()
			if !client.login {
				client.Login()
			}
			queryLock.Unlock()
		}
		resp, err = client.CallApi("/profile/get_profile", request)
		if err != nil {
			return resp, err
		}
	}

	if save {
		userInfo := resp.Get("user_info")
		gamerInfo := &GamerInfoUnit{}
		gamerInfo.TimeStamp = userInfo.Get("last_login_time").Int()
		gamerInfo.ArenaRank = userInfo.Get("arena_rank").Int()
		if gamerInfo.ArenaRank == 0 {
			log.Error().Str("name", pluginName).Msgf("%v", resp)
			return resp, nil
		}
		gamerInfo.GrandArenaRank = userInfo.Get("grand_arena_rank").Int()
		gamerInfo.GamerName = userInfo.Get("user_name").Str
		gamerInfoManage.Lock()
		defer gamerInfoManage.Unlock()
		gamerInfoManage.TmpGamerInfoMap[id] = gamerInfo
	}

	return resp, err
}

func queryAll(idList []string, threadNum int) {
	gamerInfoManage.Lock()
	gamerInfoManage.TmpGamerInfoMap = map[string]*GamerInfoUnit{}
	gamerInfoManage.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(threadNum)
	idChan := make(chan string, 100)

	for i := 0; i < threadNum; i++ {
		go func() {
			defer wg.Done()
			for id := range idChan {
				query(id, true)
			}
		}()
	}

	for _, id := range idList {
		idChan <- id
	}
	close(idChan)

	wg.Wait()
}

func judgeMode(mode string, modeStr string) bool {
	if loc, ok := modeLoc[mode]; ok {
		if len(modeStr) >= loc && modeStr[loc-1] == '1' {
			return true
		}
	}
	return false
}

func getChange(id string, num int, mode string) string {

	gamerInfoManage.RLock()
	defer gamerInfoManage.RUnlock()
	if _, ok := gamerInfoManage.GamerInfoMap[id]; !ok {
		return ""
	}
	if _, ok := gamerInfoManage.TmpGamerInfoMap[id]; !ok {
		return ""
	}

	old := gamerInfoManage.GamerInfoMap[id]
	new := gamerInfoManage.TmpGamerInfoMap[id]

	msg := ""
	var preMsg string
	if num == 0 {
		preMsg = "您"
	} else {
		preMsg = fmt.Sprintf("您的关注%d %s", num, new.GamerName)
	}
	// 上线信息处理
	if new.TimeStamp-old.TimeStamp > 1800 && judgeMode("login_remind", mode) {
		msg += "已上线"
	}

	// 竞技场处理
	if old.ArenaRank != new.ArenaRank && judgeMode("arena_on", mode) {
		if old.ArenaRank < new.ArenaRank || judgeMode("rank_up", mode) {
			msg += fmt.Sprintf(" jjc:%d->%d", old.ArenaRank, new.ArenaRank)
		}
	}

	// 公主竞技场处理
	if old.GrandArenaRank != new.GrandArenaRank && judgeMode("grand_arena_on", mode) {
		if old.GrandArenaRank < new.GrandArenaRank || judgeMode("rank_up", mode) {
			msg += fmt.Sprintf(" pjjc:%d->%d", old.GrandArenaRank, new.GrandArenaRank)
		}
	}

	if msg == "" {
		return msg
	}

	return preMsg + msg
}

func updateVersion(version string) error {
	for _, client := range clientMap {
		client.updateVersion(version)
	}
	header["APP-VER"] = version
	err := savaHeader()
	return err
}
