package pcrjjc3

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type UserInfoUnit struct {
	Id   []string `json:"id"`
	Bid  []string `json:"bid"`
	Gid  []string `json:"gid"`
	Mode []string `json:"mode"`
}

type UserInfoManage struct {
	UserInfoMap  map[string]UserInfoUnit `json:"arena_bind"`
	idList       []string                `json:"-"`
	update       bool                    `json:"-"`
	sync.RWMutex `json:"-"`
}

func (UIM *UserInfoManage) GetIdList() []string {
	UIM.Lock()
	defer UIM.Unlock()
	if !UIM.update {
		return UIM.idList
	}

	// 使用map去重
	idMap := map[string]bool{}
	for _, userInfo := range UIM.UserInfoMap {
		for _, id := range userInfo.Id {
			if id == "" {
				continue
			}
			idMap[id] = true
		}
	}

	idList := []string{}
	for k := range idMap {
		idList = append(idList, k)
	}

	UIM.idList = idList
	UIM.update = false
	return idList
}

func (UIM *UserInfoManage) bind(id string, uid string, gid string, bid string, ifAttention bool) (string, error) {
	var msg string
	var err error

	_, err = checkServer(id[:1])
	if err != nil {
		return "", err
	}

	UIM.Lock()
	defer UIM.Unlock()

	userMap := UIM.UserInfoMap
	if _, ok := userMap[uid]; !ok {
		userMap[uid] = UserInfoUnit{
			Id:   make([]string, 2),
			Bid:  make([]string, 2),
			Gid:  make([]string, 2),
			Mode: make([]string, 2),
		}
	}

	userInfo := userMap[uid]
	if ifAttention {
		if len(userInfo.Id) >= config.AccountLimit {
			msg = fmt.Sprintf("因为服务器性能有限，仅支持关注%d名玩家", config.AccountLimit-1)
			return msg, nil
		} else {
			for _, exisId := range userInfo.Id {
				if exisId == id {
					msg = "您已关注或绑定该玩家"
					return msg, nil
				}
			}
			userInfo.Id = append(userInfo.Id, id)
			userInfo.Gid = append(userInfo.Gid, gid)
			userInfo.Bid = append(userInfo.Bid, bid)
			userInfo.Mode = append(userInfo.Mode, "1111")
		}
	} else {
		userInfo.Id[0] = id
		userInfo.Bid[0] = bid
		userInfo.Gid[0] = gid
		userInfo.Mode[0] = "1100"
	}
	userMap[uid] = userInfo

	err = UIM.saveUserInfo()
	if err != nil {
		return "", err
	}

	if ifAttention {
		msg = "竞技场关注成功"
	} else {
		msg = "竞技场bind成功"
	}

	UIM.update = true

	return msg, nil
}

func (UIM *UserInfoManage) saveUserInfo() error {
	userJSON, err := json.MarshalIndent(UIM, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dataPath, "binds.json"), userJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (UIM *UserInfoManage) attentionList(uid string) string {
	UIM.RLock()
	defer UIM.RUnlock()
	if _, ok := userInfoManage.UserInfoMap[uid]; !ok {
		return "您没有关注任何玩家"
	}
	userInfo := userInfoManage.UserInfoMap[uid]
	if len(userInfo.Id) < 2 {
		return "您没有关注任何玩家"
	}

	msg := gamerInfoManage.getSnapshot(userInfo.Id[1:])
	return msg
}

func (UIM *UserInfoManage) delBind(uid string, num int) string {
	UIM.RLock()

	if _, ok := UIM.UserInfoMap[uid]; !ok {
		UIM.RUnlock()
		return "您没有绑定任何游戏账号"
	}
	userInfo := UIM.UserInfoMap[uid]
	if num >= len(userInfo.Id) {
		UIM.RUnlock()
		return "请输入正确的关注序号"
	}
	UIM.RUnlock()

	UIM.Lock()
	if num == 0 {
		userInfo.Id[0] = ""
		userInfo.Bid[0] = ""
		userInfo.Gid[0] = ""
		userInfo.Mode[0] = ""
	} else {
		userInfo.Id = slices.Delete(userInfo.Id, num, num+1)
		userInfo.Bid = slices.Delete(userInfo.Bid, num, num+1)
		userInfo.Gid = slices.Delete(userInfo.Gid, num, num+1)
		userInfo.Mode = slices.Delete(userInfo.Mode, num, num+1)
	}

	UIM.UserInfoMap[uid] = userInfo
	if len(userInfo.Id) == 1 && userInfo.Id[0] == "" {
		delete(UIM.UserInfoMap, uid)
	}
	UIM.saveUserInfo()
	UIM.update = true
	UIM.Unlock()

	return "删除成功"
}
