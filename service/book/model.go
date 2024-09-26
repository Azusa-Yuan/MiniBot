package book

import database "MiniBot/utils/db"

type Book struct {
	ID      int64  `gorm:"primaryKey"` // 自增主键
	BotID   int64  `gorm:"column:bot_id; index"`
	UserID  int64  `gorm:"column:user_id"`
	GroupID int64  `grom:"column:group_id"`
	Service string `grom:"column:service; index"`
	Value   string `gorm:"column:value"`
}

var db = database.GetDefalutDB()

func init() {
	db.AutoMigrate(&Book{})
}

func GetBookInfos(service string) ([]Book, error) {
	bookInfos := []Book{}
	err := db.Where("service = ?", service).Find(&bookInfos).Error
	return bookInfos, err
}

func CreatOrUpdateBookInfo(bookInfo *Book) error {
	tx := db.Begin()

	oldInfo := Book{}
	tx.Where("bot_id = ? AND user_id = ? AND group_id = ? AND service = ?", bookInfo.BotID, bookInfo.UserID, bookInfo.GroupID, bookInfo.Service).
		Find(&oldInfo)
	if oldInfo.ID > 0 {
		bookInfo.ID = oldInfo.ID
	}

	err := db.Save(&bookInfo).Error
	return err
}
