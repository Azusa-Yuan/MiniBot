package bilibili

import (
	database "MiniBot/utils/db"
	"MiniBot/utils/net_tools"

	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

var (
	vtbURLs = [...]string{"https://api.vtbs.moe/v1/short"}
	vdb     *vupdb
)

// vupdb 分数数据库
type vupdb gorm.DB

type vup struct {
	Mid    int64  `gorm:"column:mid;primary_key"`
	Uname  string `gorm:"column:uname"`
	Roomid int64  `gorm:"column:roomid"`
}

func (vup) TableName() string {
	return "vup"
}

// initializeVup 初始化vup数据库
func initializeVup() (*vupdb, error) {
	gdb := database.GetDefalutDB()
	gdb.AutoMigrate(&vup{})
	return (*vupdb)(gdb), nil
}

func (vdb *vupdb) insertVupByMid(mid int64, uname string, roomid int64) (err error) {
	db := (*gorm.DB)(vdb)
	v := vup{
		Mid:    mid,
		Uname:  uname,
		Roomid: roomid,
	}
	err = db.Save(&v).Error
	return
}

// filterVup 筛选vup
func (vdb *vupdb) filterVup(ids []int64) (vups []vup, err error) {
	db := (*gorm.DB)(vdb)
	if err = db.Model(&vup{}).Find(&vups, "mid in (?)", ids).Error; err != nil {
		return vups, err
	}
	return
}

func updateVup() error {
	for _, v := range vtbURLs {
		data, err := net_tools.Download(v)
		if err != nil {
			return err
		}
		gjson.GetBytes(data, "@this").ForEach(func(_, value gjson.Result) bool {
			mid := value.Get("mid").Int()
			uname := value.Get("uname").String()
			roomid := value.Get("roomid").Int()
			err = vdb.insertVupByMid(mid, uname, roomid)
			return err == nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
