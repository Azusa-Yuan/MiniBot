package qqwife

import (
	database "MiniBot/utils/db"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMo(t *testing.T) {

	dsn := "host=127.0.0.1 user=postgres password=a123456 dbname=lulumu port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}
	qqwife.db = db
	qqwife.db.AutoMigrate(&GroupInfo{}, &CdSheet{}, &Favorability{}, &PigInfo{}, &UserInfo{})
}

func TestMigrate(t *testing.T) {
	db := database.DbConfig.GetDb("lulumu")
	// 删表
	db.Migrator().DropTable(&GroupInfo{}, &CdSheet{}, &Favorability{}, &PigInfo{}, &UserInfo{})
	db.AutoMigrate(&GroupInfo{}, &CdSheet{}, &Favorability{}, &PigInfo{}, &UserInfo{})
	db2, _ := sql.Open("sqlite3", "./qqwife.db")
	// 查询数据
	rows, err := db2.Query("SELECT Userinfo, Favor FROM favorability")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// 输出查询结果
	for rows.Next() {
		var userInfo string
		var favor int
		if err := rows.Scan(&userInfo, &favor); err != nil {
			log.Fatal(err)
		}
		pair := strings.Split(userInfo, "+")

		qq1, _ := strconv.ParseInt(pair[0], 10, 64)
		qq2, _ := strconv.ParseInt(pair[1], 10, 64)
		db.Create(&Favorability{
			UID: qq1, Target: qq2, Favor: favor,
		})
		db.Create(&Favorability{
			UID: qq2, Target: qq1, Favor: favor,
		})
		fmt.Printf(" %v,  %d, \n", pair, favor)
	}
}
