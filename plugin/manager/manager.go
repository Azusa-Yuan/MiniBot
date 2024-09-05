// Package control 控制插件的启用与优先级等
package manager

import (
	database "MiniBot/utils/db"
	"sync"

	"github.com/sirupsen/logrus"
)

// Manager 管理
type Manager struct {
	sync.RWMutex
}

// 针对全部插件的
var (
	// 封禁uid的set
	blockCache = make(map[string]bool)
	// 沉默的gid的set
	db         = database.DbConfig.GetDb("lulumu")
	LevelCache = make(map[string]uint)
)

// NewManager
func NewManager() (m Manager) {
	db.AutoMigrate(&BanUser{}, &ClosePlugin{}, &PermissionLevel{})
	m = Manager{}
	m.Lock()
	defer m.Unlock()

	banUsers := []BanUser{}
	db.Select("key").Find(&banUsers)
	for _, banUser := range banUsers {
		blockCache[banUser.Key] = true
	}

	permissionLevels := []PermissionLevel{}
	db.Select("key", "level").Find(&permissionLevels)
	logrus.Debug("权限查询结构", permissionLevels)
	for _, permissionLevel := range permissionLevels {
		LevelCache[permissionLevel.Key] = permissionLevel.Level
	}
	logrus.Debug("权限等级缓存", LevelCache)
	return
}

func (manager *Manager) GetLevel(key string) uint {
	manager.RLock()
	defer manager.RUnlock()
	return LevelCache[key]
}

func (manager *Manager) SetLevel(key string, level uint) {
	manager.Lock()
	defer manager.Unlock()
	LevelCache[key] = level
	keyLevel := PermissionLevel{Key: key, Level: level}
	db.Where(PermissionLevel{Key: key}).Assign(keyLevel).FirstOrCreate(&keyLevel)
}

func (manager *Manager) DoBlock(key string) {
	manager.Lock()
	defer manager.Unlock()
	if _, ok := blockCache[key]; ok {
		return
	}
	blockCache[key] = true
	db.Create(&BanUser{Key: key})
}

// DoUnblock 解封
func (manager *Manager) DoUnblock(key string) {
	manager.Lock()
	defer manager.Unlock()
	if _, ok := blockCache[key]; !ok {
		return
	}
	delete(blockCache, key)
	db.Where("key = ?", key).Delete(&BanUser{})
}

// IsBlocked 是否封禁
func (manager *Manager) IsBlocked(key string) bool {
	manager.RLock()
	defer manager.RUnlock()
	_, ok := blockCache[key]
	return ok
}

// Lookup returns a Manager by the service name, if
// not exist, it will return nil.
