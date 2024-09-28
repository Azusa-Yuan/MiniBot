package eqa

type eqa struct {
	ID    int64  `gorm:"primaryKey"` // 自增主键
	Key   string `gorm:"column:key; index"`
	Value string `gorm:"value"`
}
