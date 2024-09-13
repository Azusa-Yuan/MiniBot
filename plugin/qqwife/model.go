package qqwife

import (
	"errors"
	"sync"
	"time"

	database "MiniBot/utils/db"

	"gorm.io/gorm"
)

type QQWife struct {
	db *gorm.DB
	sync.RWMutex
}

// sync.RWMutex 是一个结构体类型，用于实现读写锁。在定义结构体时，如果将 sync.RWMutex 直接作为结构体的匿名字段（即不加变量名），
// 是为了在结构体中继承 sync.RWMutex 的方法和功能，而不需要显式地声明一个字段名。
var (
	qqwife = &QQWife{}
	db     *gorm.DB
)

func init() {
	db = database.DbConfig.GetDb("lulumu")
	qqwife.db = db
	err := db.AutoMigrate(&GroupInfo{}, &CdSheet{}, &Favorability{}, &PigInfo{}, &UserInfo{})
	if err != nil {
		return
	}

}

// 定义 SQLite 数据模型
type Favorability struct {
	UID    int64 `gorm:"column:uid; primaryKey"`
	Target int64 `gorm:"column:target; primaryKey"` // 记录用户
	Favor  int   // 好感度
}

// 群设置
type GroupInfo struct {
	GID        int64   `gorm:"column:gid; primaryKey"`
	CanMatch   int     // 嫁婚开关
	CanNtr     int     // Ntr开关
	CDtime     float64 // CD时间
	Updatetime string  // 登记时间
}

// 猪头信息
type PigInfo struct {
	ID         int64  `gorm:"column:id; primaryKey"`
	GID        int64  `gorm:"column:gid; index"`
	UID        int64  `gorm:"column:uid"`
	Updatetime string // 登记时间
}

// 结婚信息
type UserInfo struct {
	ID         int64     `gorm:"column:id; primaryKey"`
	GID        int64     `gorm:"column:gid; unique_index:gid_uid; not null; unique_index:gid_target"`
	UID        int64     `gorm:"column:uid; unique_index:gid_uid;not null"`
	Target     int64     `gorm:"column:target; unique_index:gid_target"` // 对象身份证号
	Username   string    `gorm:"column:user_name"`                       // 户主名称
	Targetname string    `gorm:"column:target_name"`                     // 对象名称
	Updatetime time.Time // 登记时间
	Mode       uint16    `gorm:"column:mode"`
}

// cd信息
type CdSheet struct {
	ID     int64  `gorm:"column:id; primaryKey"`
	Time   int64  // 时间
	GID    int64  `gorm:"column:gid"` // 群号
	UID    int64  `gorm:"column:uid"` // 用户
	ModeID string // 技能类型
}

// 返回今日猪头的信息
func (sql *QQWife) GetPigs(gid int64) (pigsInfos []PigInfo, err error) {
	pigInfo := PigInfo{}
	res := sql.db.First(&pigInfo)
	if res.RowsAffected != 0 {
		if time.Now().Format("2006/01/02") != pigInfo.Updatetime {
			// 如果跨天了就删除
			err = sql.db.Migrator().DropTable(&PigInfo{})
			if err != nil {
				return
			}
			err = sql.db.Migrator().CreateTable(&PigInfo{})
			if err != nil {
				return
			}
		}
		return
	}

	res = sql.db.Where("gid = ?", gid).Find(&pigsInfos)
	if res.Error != nil {
		err = res.Error
	}
	return
}

// 保存今日猪头的信息
func (sql *QQWife) SavePigs(gid int64, pigs []int64) error {
	pigInfos := []PigInfo{}
	for _, pig := range pigs {
		pigInfo := PigInfo{
			GID:        gid,
			UID:        pig,
			Updatetime: time.Now().Format("2006/01/02")}
		pigInfos = append(pigInfos, pigInfo)
	}

	res := sql.db.Create(&pigInfos)
	return res.Error
}

// 返回今日老婆的群设置
func (sql *QQWife) GetGroupInfo(gid int64) (dbinfo GroupInfo, err error) {
	res := sql.db.Where("gid = ?", gid).First(&dbinfo)
	if res.Error != nil {
		dbinfo = GroupInfo{
			GID:        gid,
			CanMatch:   1,
			CanNtr:     1,
			CDtime:     1,
			Updatetime: time.Now().Format("2006/01/02"),
		}
		db.Create(&dbinfo)
		// 没有记录
		return dbinfo, nil
	}
	return
}

func (sql *QQWife) UpdateGroupInfo(dbinfo GroupInfo) error {
	res := sql.db.Where("gid = ?", dbinfo.GID).Assign(dbinfo).FirstOrCreate(&GroupInfo{})
	return res.Error
}

func (sql *QQWife) IfToday() (err error) {
	userInfo := &UserInfo{}
	res := sql.db.First(&userInfo)
	if res.RowsAffected == 0 {
		return
	}

	if time.Now().Format("2006/01/02") != userInfo.Updatetime.Format("2006/01/02") {
		// 如果跨天了就删除
		err = sql.db.Migrator().DropTable(&UserInfo{})
		if err != nil {
			return
		}
		err = sql.db.Migrator().CreateTable(&UserInfo{})
		if err != nil {
			return
		}
	}
	return
}

func (sql *QQWife) GetMarriageInfo(gid, uid int64) (UserInfo, error) {
	info := UserInfo{}
	err := sql.db.Where("uid = ? AND gid = ?", uid, gid).First(&info).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return info, err
	}
	return info, nil
}

// 民政局登记数据
func (sql *QQWife) SaveMarriageInfo(gid, husband, wife int64, husbandname, wifetname string) error {
	uidinfo := UserInfo{
		GID:        gid,
		UID:        husband,
		Username:   husbandname,
		Target:     wife,
		Targetname: wifetname,
		Updatetime: time.Now(),
		Mode:       1,
	}
	sql.db.Create(&uidinfo)
	uidinfo = UserInfo{
		GID:        gid,
		UID:        wife,
		Username:   wifetname,
		Target:     husband,
		Targetname: husbandname,
		Updatetime: time.Now(),
		Mode:       0,
	}
	res := sql.db.Create(&uidinfo)
	return res.Error
}

func (sql *QQWife) GetAllInfo(gid int64) (list []UserInfo, err error) {
	res := sql.db.Where("gid = ? AND mode = 1", gid).Find(&list)
	if res.Error != nil {
		err = res.Error
	}
	return
	// for _, info := range infos {
	// 	dbinfo := [4]string{
	// 		info.Username,
	// 		strconv.FormatInt(info.UID, 10),
	// 		info.Targetname,
	// 		strconv.FormatInt(info.Target, 10),
	// 	}
	// 	list = append(list, dbinfo)
	// }
	// return
}

// func (sql *QQWife) 清理花名册(gid ...string) error {
// 	sql.Lock()
// 	defer sql.Unlock()
// 	switch gid {
// 	case nil:
// 		grouplist, err := sql.db.ListTables()
// 		if err == nil {
// 			for _, listName := range grouplist {
// 				if listName == "Favorability" {
// 					continue
// 				}
// 				err = sql.db.Drop(listName)
// 			}
// 		}
// 		return err
// 	default:
// 		err := sql.db.Drop(gid[0])
// 		if err == nil {
// 			_ = sql.db.Del("CdSheet", "where GroupID is "+strings.ReplaceAll(gid[0], "group", ""))
// 		}
// 		return err
// 	}
// }

func (sql *QQWife) JudgeCD(gid, uid int64, mode string, cdtime float64) (ok bool, err error) {

	cdinfo := CdSheet{}
	res := sql.db.Where("gid = ? AND uid = ? AND mode_id = ?", gid, uid, mode).First(&cdinfo)
	if res.RowsAffected == 0 {
		// 没有记录即不用比较
		return true, nil
	}

	if time.Since(time.Unix(cdinfo.Time, 0)).Hours() > cdtime {
		// 如果CD已过就删除
		err = sql.db.Delete(&cdinfo).Error
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (sql *QQWife) SaveCD(gid, uid int64, mode string) error {
	res := sql.db.Create(&CdSheet{
		Time:   time.Now().Unix(),
		GID:    gid,
		UID:    uid,
		ModeID: mode,
	})
	return res.Error
}

// func (sql *QQWife) DelWifeInfo(gid, wife int64) error {
// 	res := sql.db.Where("target = ? AND gid = ?", wife, gid).Delete(&UserInfo{})
// 	return res.Error
// }

// func (sql *QQWife) DelHisbandInfo(gid, husband int64) error {
// 	res := sql.db.Where("uid = ? AND gid = ?", husband, gid).Delete(&UserInfo{})
// 	return res.Error
// }

func (sql *QQWife) DelMarriageInfo(gid, uid int64) (err error) {
	err = sql.db.Where("uid = ? AND gid = ?", uid, gid).Delete(&UserInfo{}).Error
	if err != nil {
		return
	}

	err = sql.db.Where("target = ? AND gid = ?", uid, gid).Delete(&UserInfo{}).Error
	return
}

func (sql *QQWife) UpdateFavorability(uid, target int64, score int) (favor int, err error) {
	info := Favorability{}
	info.UID = uid
	info.Target = target
	err = sql.db.Where("uid = ? AND target = ?", uid, target).First(&info).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
	}
	info.Favor += score
	if info.Favor > 100 {
		info.Favor = 100
	} else if info.Favor < 0 {
		info.Favor = 0
	}
	db.Save(&info)

	// 需要双向存储
	info.UID = target
	info.Target = uid
	db.Save(&info)
	favor = info.Favor
	return
}

func (sql *QQWife) GetFavorability(uid, target int64) (int, error) {
	info := Favorability{}
	res := sql.db.Where("uid = ? And target = ?", uid, target).First(&info)
	if res.Error != nil {
		return 0, nil
	}
	return info.Favor, nil
}

func (sql *QQWife) GetFavorabilityList(uid int64) (list []Favorability, err error) {
	res := sql.db.Where("uid = ?", uid).Find(&list)
	if res.Error != nil {
		err = res.Error
		return
	}
	return
}
