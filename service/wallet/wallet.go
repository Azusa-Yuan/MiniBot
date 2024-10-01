// Package wallet 货币系统
package wallet

import (
	zero "ZeroBot"
	"sync"

	database "MiniBot/utils/db"
	"MiniBot/utils/transform"

	"gorm.io/gorm"
)

// Storage 货币系统
type Storage struct {
	sync.RWMutex
	db *gorm.DB
}

// Wallet 钱包
type Wallet struct {
	UID   int64 `gorm:"index;column:uid;unique"`
	Money int   `gorm:"column:money"`
}

var (
	sdb = &Storage{
		db: database.GetDefalutDB(),
	}
)

func init() {
	sdb.db.AutoMigrate(&Wallet{})
}

// GetWalletOf 获取钱包数据
func GetWalletMoneyByCtx(ctx *zero.Ctx) (money int) {
	uid := transform.BidWithuidInt64(ctx)
	return sdb.getWalletOf(uid).Money
}

// GetGroupWalletOf 获取多人钱包数据
//
// if sort == true,由高到低排序; if sort == false,由低到高排序 todo
func GetGroupWalletOf(sortable bool, bid int64, uids ...int64) (wallets []Wallet, err error) {
	return sdb.getGroupWalletOf(sortable, uids...)
}

// UpdateWalletOf 更新钱包(money > 0 增加,money < 0 减少)
func UpdateWalletByCtx(ctx *zero.Ctx, money int) error {
	uid := transform.BidWithuidInt64(ctx)
	return sdb.updateWalletWithAdd(uid, money)
}

// 获取钱包数据
func (sql *Storage) getWalletOf(uid int64) (wallet Wallet) {
	sql.db.Where("uid = ?", uid).First(&wallet)
	return
}

// 获取钱包数据组
func (sql *Storage) getGroupWalletOf(sortable bool, uids ...int64) (wallets []Wallet, err error) {
	wallets = make([]Wallet, 0, len(uids))
	sort := "ASC"
	if sortable {
		sort = "DESC"
	}
	sql.db.Where("uid IN ?", uids).Order("money " + sort).Find(&wallets)
	return
}

// // 更新钱包
// func (sql *Storage) updateWalletOf(uid int64, money int) (err error) {
// 	sql.Lock()
// 	defer sql.Unlock()
// 	return sql.db.Insert("storage", &Wallet{
// 		UID:   uid,
// 		Money: money,
// 	})
// }

// 更新钱包
func (sql *Storage) updateWalletWithAdd(uid int64, money int) (err error) {
	wallet := Wallet{UID: uid, Money: money}
	if sql.db.Where("uid = ?", uid).First(&wallet).RowsAffected > 0 {
		err = sql.db.Model(&Wallet{}).Where("uid = ?", uid).Update("money", gorm.Expr("money + ?", money)).Error
	} else {
		sql.db.Create(&wallet)
	}
	return

}
