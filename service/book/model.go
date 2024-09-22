package book

type book struct {
	BotID   int64  `gorm:"column:bot_id"`
	UserID  int64  `gorm:"column:user_id"`
	GroupID int64  `grom:"column:group_id"`
	Service string `grom:"column:service"`
	Value   string `gorm:"column:value"`
}
