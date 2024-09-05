package manager

import (
	"strconv"

	"gorm.io/gorm"
)

// GroupConfig holds the group config for the Manager.
type GroupConfig struct {
	GroupID int64 `db:"gid"`     // GroupID 群号
	Disable int64 `db:"disable"` // Disable 默认启用该插件
}

// BanStatus 在某群封禁某人的状态
type BanStatus struct {
	ID      int64 `db:"id"`
	UserID  int64 `db:"uid"`
	GroupID int64 `db:"gid"`
}

// BlockStatus 全局 ban 某人
type BlockStatus struct {
	UserID int64 `db:"uid"`
}

// ResponseGroup 响应的群
type ResponseGroup struct {
	GroupID int64  `db:"gid"` // GroupID 群号, 个人为负
	Extra   string `db:"ext"` // Extra 该群的扩展数据
}

// Options holds the optional parameters for the Manager.

func bidWithuid(bid, uid int64) string {
	return strconv.FormatInt(bid, 10) + "_uid" + strconv.FormatInt(uid, 10)
}

func bidWithgid(bid, gid int64) string {
	return strconv.FormatInt(bid, 10) + "_gid" + strconv.FormatInt(gid, 10)
}

type Plugin struct {
	gorm.Model
	Name  string `gorm:"index;column:name;type:varchar(50)"`
	Count int64  `gorm:"column:count"`
}

type BanUser struct {
	gorm.Model
	Bid int64  `gorm:"index;column:bid"`
	UID int64  `gorm:"column:uid"`
	Key string `gorm:"uniqueIndex;column:key;type:varchar(50)"`
}

// type CloseGroup struct {
// 	gorm.Model
// 	Bid int64  `gorm:"index;column:bid"`
// 	GID int64  `gorm:"column:gid"`
// 	Key string `gorm:"uniqueIndex;column:key;type:varchar(50)"`
// }

type ClosePlugin struct {
	gorm.Model
	Bid        int64  `gorm:"index;column:bid"`
	GID        int64  `gorm:"column:gid"`
	Key        string `gorm:"index;column:key;type:varchar(50)"`
	PluginName string `gorm:"index;column:plugin_name;type:varchar(50)"`
}

type PermissionLevel struct {
	gorm.Model
	Key   string `gorm:"index;column:key;type:varchar(50)"`
	Level uint   `gorm:"column:level"`
}
