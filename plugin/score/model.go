package score

import (
	database "MiniBot/utils/db"
	"time"

	"gorm.io/gorm"
)

var pluginName = "score"

// sdb 得分数据库
var sdb *scoredb

// scoredb 分数数据库
type scoredb gorm.DB

// scoretable 分数结构体
type scoretable struct {
	UID   int64 `gorm:"column:uid;primaryKey"`
	Score int   `gorm:"column:score;default:0"`
}

// TableName ...
func (scoretable) TableName() string {
	return "score"
}

// signintable 签到结构体
type signintable struct {
	UID       int64 `gorm:"column:uid;primaryKey"`
	Count     int   `gorm:"column:count;default:0"`
	UpdatedAt time.Time
}

// TableName ...
func (signintable) TableName() string {
	return "sign_in"
}

// initialize 初始化ScoreDB数据库
func initialize() *scoredb {
	gdb := database.DbConfig.GetDb("lulumu")
	gdb.AutoMigrate(&scoretable{}, &signintable{})
	return (*scoredb)(gdb)
}

// GetScoreByUID 取得分数
func (sdb *scoredb) GetScoreByUID(uid int64) (s scoretable) {
	db := (*gorm.DB)(sdb)
	db.Model(&scoretable{}).FirstOrCreate(&s, "uid = ? ", uid)
	return s
}

// InsertOrUpdateScoreByUID 插入或更新分数
func (sdb *scoredb) InsertOrUpdateScoreByUID(uid int64, score int) (err error) {
	db := (*gorm.DB)(sdb)
	s := scoretable{
		UID:   uid,
		Score: score,
	}
	err = db.Model(&scoretable{}).Where("uid = ?", uid).Assign(scoretable{Score: score}).FirstOrCreate(&s).Error
	return
}

// GetSignInByUID 取得签到次数
func (sdb *scoredb) GetSignInByUID(uid int64) (si signintable) {
	db := (*gorm.DB)(sdb)
	db.Model(&signintable{}).FirstOrCreate(&si, "uid = ? ", uid)
	return si
}

// ResetTable 重置签到表
func (sdb *scoredb) ResetTable() error {
	db := (*gorm.DB)(sdb)
	err := db.Migrator().DropTable(&signintable{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&signintable{})
	if err != nil {
		return err
	}
	return nil
}

// InsertOrUpdateSignInCountByUID 插入或更新签到次数
func (sdb *scoredb) InsertOrUpdateSignInCountByUID(uid int64, count int) (err error) {
	db := (*gorm.DB)(sdb)
	si := signintable{
		UID:   uid,
		Count: count,
	}
	err = db.Where(signintable{UID: uid}).Assign(signintable{Count: count}).FirstOrCreate(&si).Error
	return
}

func (sdb *scoredb) GetScoreRankByTopN(n int) (st []scoretable, err error) {
	db := (*gorm.DB)(sdb)
	err = db.Model(&scoretable{}).Order("score desc").Limit(n).Find(&st).Error
	return
}

type scdata struct {
	drawedfile string
	picfile    string
	uid        int64
	nickname   string
	inc        int // 增加币
	score      int // 钱包
	level      int
	rank       int
}
