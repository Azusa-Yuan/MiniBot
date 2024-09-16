package sleepmanage

import (
	database "MiniBot/utils/db"
	"time"

	"gorm.io/gorm"
)

// sdb 睡眠数据库全局变量
var sdb *sleepdb

// sleepdb 睡眠数据库结构体
type sleepdb gorm.DB

// initialize 初始化
func initialize() *sleepdb {

	gdb := database.DbConfig.GetDb("lulumu")
	gdb.AutoMigrate(&SleepManage{})
	return (*sleepdb)(gdb)
}

// SleepManage 睡眠信息
type SleepManage struct {
	ID        uint      `gorm:"primary_key"`
	GroupID   int64     `gorm:"column:group_id"`
	UserID    int64     `gorm:"column:user_id"`
	SleepTime time.Time `gorm:"column:sleep_time"`
}

// TableName 表名
func (SleepManage) TableName() string {
	return "sleep_manage"
}

// sleep 更新睡眠时间
func (sdb *sleepdb) sleep(gid, uid int64) (position int64, awakeTime time.Duration) {
	db := (*gorm.DB)(sdb)
	now := time.Now()
	var today time.Time
	if now.Hour() >= 21 {
		today = now.Add(-time.Hour*time.Duration(-21+now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second()))
	} else if now.Hour() <= 3 {
		today = now.Add(-time.Hour*time.Duration(3+now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second()))
	}
	st := SleepManage{
		GroupID:   gid,
		UserID:    uid,
		SleepTime: now,
	}
	if err := db.Model(&SleepManage{}).Where("group_id = ? and user_id = ?", gid, uid).First(&st).Error; err != nil {
		// error handling...
		if err == gorm.ErrRecordNotFound {
			db.Model(&SleepManage{}).Create(&st) // newUser not user
		}
	} else {
		awakeTime = now.Sub(st.SleepTime)
		db.Model(&SleepManage{}).Where("group_id = ? and user_id = ?", gid, uid).Updates(
			map[string]any{
				"sleep_time": now,
			})
	}
	db.Model(&SleepManage{}).Where("group_id = ? and sleep_time <= ? and sleep_time >= ?", gid, now, today).Count(&position)
	return position, awakeTime
}

// getUp 更新起床时间
func (sdb *sleepdb) getUp(gid, uid int64) (position int64, sleepTime time.Duration) {
	db := (*gorm.DB)(sdb)
	now := time.Now()
	today := now.Add(-time.Hour*time.Duration(-6+now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second()))
	st := SleepManage{
		GroupID:   gid,
		UserID:    uid,
		SleepTime: now,
	}
	if err := db.Model(&SleepManage{}).Where("group_id = ? and user_id = ?", gid, uid).First(&st).Error; err != nil {
		// error handling...
		if err == gorm.ErrRecordNotFound {
			db.Model(&SleepManage{}).Create(&st) // newUser not user
		}
	} else {
		sleepTime = now.Sub(st.SleepTime)
		db.Model(&SleepManage{}).Where("group_id = ? and user_id = ?", gid, uid).Updates(
			map[string]any{
				"sleep_time": now,
			})
	}
	db.Model(&SleepManage{}).Where("group_id = ? and sleep_time <= ? and sleep_time >= ?", gid, now, today).Count(&position)
	return position, sleepTime
}
