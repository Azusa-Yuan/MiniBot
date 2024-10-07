package book

import database "MiniBot/utils/db"

type Book struct {
	ID      int64  `gorm:"primaryKey"` // 自增主键
	BotID   int64  `gorm:"column:bot_id; index"`
	UserID  int64  `gorm:"column:user_id"`
	GroupID int64  `gorm:"column:group_id"`
	Service string `gorm:"column:service; index"`
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

func GetUserBookInfo(bookInfo *Book) (*Book, error) {
	err := db.Where("bot_id = ? AND user_id = ? AND group_id = ? AND service = ?", bookInfo.BotID, bookInfo.UserID, bookInfo.GroupID, bookInfo.Service).
		Find(&bookInfo).Error
	return bookInfo, err
}

func CreatOrUpdateBookInfo(bookInfo *Book) error {
	tx := db.Begin()

	oldInfo := Book{}
	err := tx.Where("bot_id = ? AND user_id = ? AND group_id = ? AND service = ?", bookInfo.BotID, bookInfo.UserID, bookInfo.GroupID, bookInfo.Service).
		Find(&oldInfo).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if oldInfo.ID > 0 {
		bookInfo.ID = oldInfo.ID
	}

	err = tx.Save(&bookInfo).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func DeleteBookInfo(bookInfo *Book) error {
	tx := db.Begin()

	oldInfo := Book{}
	err := tx.Where("bot_id = ? AND user_id = ? AND group_id = ? AND service = ?", bookInfo.BotID, bookInfo.UserID, bookInfo.GroupID, bookInfo.Service).
		Delete(&oldInfo).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func DeleteBookInfoByID(id int64) error {
	tx := db.Begin()

	oldInfo := Book{}
	err := tx.Where("id = ?", id).Delete(&oldInfo).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
