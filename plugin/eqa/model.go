package eqa

import "ZeroBot/message"

type eqa struct {
	ID          int64           `gorm:"primaryKey"` // 自增主键
	Key         string          `gorm:"column:key; index"`
	Value       string          `gorm:"value"`
	GID         int64           `gorm:"column:gid"`
	MessageList message.Message `gorm:"-"`
}
