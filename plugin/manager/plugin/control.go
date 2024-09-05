package plugin

import (
	database "MiniBot/utils/db"
	zero "ZeroBot"
	"sync"

	"gorm.io/gorm"
)

var db = database.DbConfig.GetDb("lulumu")

type ClosePlugin struct {
	gorm.Model
	BID        int64  `gorm:"index;column:bid"`
	GID        int64  `gorm:"column:gid"`
	Key        string `gorm:"index;column:key;type:varchar(50)"`
	PluginName string `gorm:"index;column:plugin_name;type:varchar(50)"`
}

type Options struct {
	DisableOnDefault bool
	Extra            int16  // 插件申请的 Extra 记录号, 可为 -32768~32767, 0 不可用
	Brief            string // 简介
}

// ControlManger
var CM = &ControlManger{ControlMap: map[string]*Control{}}

type ControlManger struct {
	ControlMap map[string]*Control
	sync.RWMutex
}

// Control is to control the plugins.
type Control struct {
	MetaDate *zero.MetaData
	Cache    map[string]bool // map[gid]isdisable
	Options  Options
	sync.RWMutex
}

func NewControl(metaDate *zero.MetaData) *Control {
	control := &Control{
		MetaDate: metaDate,
		Cache:    make(map[string]bool, 16),
	}

	control.Lock()
	defer control.Unlock()

	closePlugins := []ClosePlugin{}
	db.Select("key").Where("plugin_name = ?", metaDate.Name).Find(&closePlugins)
	for _, v := range closePlugins {
		control.Cache[v.Key] = true
	}
	return control
}

// Enable enables a group to pass the Manager.
// groupID == 0 (ALL) will operate on all grps.
func (c *Control) Enable(key string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.Cache[key]; !ok {
		return
	}

	db.Where("plugin_name = ? And key = ?", c.MetaDate.Name, key).Delete(&ClosePlugin{})
	delete(c.Cache, key)
}

func (c *Control) Disable(key string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.Cache[key]; ok {
		return
	}

	db.Create(&ClosePlugin{PluginName: c.MetaDate.Name,
		Key: key})
	c.Cache[key] = true
}

// IsEnabledIn 查询开启群或者插件是否在该群关闭
// 当全局未配置或与默认相同时, 状态取决于单独配置, 后备为默认配置；
// 当全局与默认不同时, 状态取决于全局配置, 单独配置失效。
func (m *Control) IsEnabled(key string) bool {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.Cache[key]; ok {
		return false
	}
	return true
}

// String 打印帮助
func (m *Control) String() string {
	return m.MetaDate.Help
}

func (cm *ControlManger) NewControl(metaDate *zero.MetaData) *Control {
	control := NewControl(metaDate)
	cm.Lock()
	defer cm.Unlock()

	cm.ControlMap[metaDate.Name] = control

	return control
}

// ForEach iterates through managers.
func (cm *ControlManger) ForEach(iterator func(key string, manager *Control) bool) {
	cm.RLock()
	m := cpmp(cm.ControlMap)
	cm.RUnlock()
	for k, v := range m {
		if !iterator(k, v) {
			return
		}
	}
}

func cpmp(m map[string]*Control) map[string]*Control {
	ret := make(map[string]*Control, len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func (cm *ControlManger) Lookup(service string) (*Control, bool) {
	cm.RLock()
	defer cm.RUnlock()
	m, ok := cm.ControlMap[service]
	return m, ok
}
