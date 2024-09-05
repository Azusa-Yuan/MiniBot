package wallet

import (
	database "MiniBot/utils/db"
	"MiniBot/utils/transform"
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestWallet(t *testing.T) {
	db := database.DbConfig.GetDb("lulumu")
	// 删表 重建
	db.Migrator().DropTable(&Wallet{})
	db.AutoMigrate(&Wallet{})
	db2, _ := sql.Open("sqlite3", "./wallet.db")
	// 查询数据
	rows, err := db2.Query("SELECT UID, Money FROM storage")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var uid int64
		var money int64
		if err := rows.Scan(&uid, &money); err != nil {
			log.Fatal(err)
		}
		db.Create(&Wallet{
			UID: transform.BidWithidInt64Custom(741433361, uid), Money: int(money),
		})
		fmt.Printf(" %v,  %d, \n", uid, money)
	}
}
