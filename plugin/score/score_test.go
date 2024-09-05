package score

import (
	database "MiniBot/utils/db"
	"MiniBot/utils/tests"
	"MiniBot/utils/transform"
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSignIn(t *testing.T) {
	mockClient := tests.CreatMockClient()
	msg := "签到"
	mockClient.Send(msg)

}

func TestMigra(t *testing.T) {
	db := database.DbConfig.GetDb("lulumu")
	// 删表 重建
	db.Migrator().DropTable(&scoretable{})
	db.AutoMigrate(&scoretable{})
	db2, _ := sql.Open("sqlite3", "./score.db")
	// 查询数据
	rows, err := db2.Query("SELECT uid, score FROM score")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var uid int64
		var score int64
		if err := rows.Scan(&uid, &score); err != nil {
			log.Fatal(err)
		}
		db.Create(&scoretable{
			UID: transform.BidWithidInt64Custom(741433361, uid), Score: int(score),
		})
		fmt.Printf(" %v,  %d, \n", uid, score)
	}
}
