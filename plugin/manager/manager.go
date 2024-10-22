// Package control 控制插件的启用与优先级等
package manager

import (
	database "MiniBot/utils/db"
	"sync"

	"github.com/rs/zerolog/log"
)

// Manager 管理
type Manager struct {
	sync.RWMutex
}

// 针对全部插件的
var (
	pluginName = "default"
	// 封禁uid的set
	blockCache = make(map[string]bool)
	// 沉默的gid的set
	db         = database.GetDefalutDB()
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
	log.Debug().Str("name", pluginName).Msgf("权限查询结构%v", permissionLevels)
	for _, permissionLevel := range permissionLevels {
		LevelCache[permissionLevel.Key] = permissionLevel.Level
	}
	log.Debug().Str("name", pluginName).Msgf("权限等级缓存%v", LevelCache)
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

func (manager *Manager) DoBlock(key string) error {
	manager.Lock()
	defer manager.Unlock()
	if _, ok := blockCache[key]; ok {
		return nil
	}
	blockCache[key] = true
	return db.Create(&BanUser{Key: key}).Error
}

// DoUnblock 解封
func (manager *Manager) DoUnblock(key string) error {
	manager.Lock()
	defer manager.Unlock()
	if _, ok := blockCache[key]; !ok {
		return nil
	}
	delete(blockCache, key)
	return db.Where("key = ?", key).Delete(&BanUser{}).Error
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
