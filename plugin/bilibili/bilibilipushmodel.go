package bilibili

import (
	database "MiniBot/utils/db"
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// bilibilipushdb bilibili推送数据库
type bilibilipushdb gorm.DB

type bilibilipush struct {
	ID             int64 `gorm:"column:id;primary_key" json:"id"`
	BilibiliUID    int64 `gorm:"column:bilibili_uid;index:idx_buid_gid" json:"bilibili_uid"`
	GroupID        int64 `gorm:"column:group_id;index:idx_buid_gid" json:"group_id"`
	BotID          int64 `gorm:"column:bot_id;index:idx_buid_gid" json:"bot_id"`
	LiveDisable    bool  `gorm:"column:live_disable;default:0" json:"live_disable"`
	DynamicDisable bool  `gorm:"column:dynamic_disable;default:0" json:"dynamic_disable"`
}

// TableName ...
func (bilibilipush) TableName() string {
	return "bilibili_push"
}

type bilibiliup struct {
	BilibiliUID int64  `gorm:"column:bilibili_uid;primary_key"`
	Name        string `gorm:"column:name"`
}

// TableName ...
func (bilibiliup) TableName() string {
	return "bilibili_up"
}

type bilibiliAt struct {
	GroupID int64 `gorm:"column:group_id;primary_key" json:"group_id"`
	AtAll   int   `gorm:"column:at_all;default:0" json:"at_all"`
}

func (bilibiliAt) TableName() string {
	return "bilibili_at"
}

// initializePush 初始化bilibilipushdb数据库
func initializePush(dbpath string) *bilibilipushdb {
	var err error
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil
		}
		defer f.Close()
	}
	gdb := database.GetDefalutDB()

	gdb.AutoMigrate(&bilibilipush{}, &bilibiliup{}, &bilibiliAt{})
	return (*bilibilipushdb)(gdb)
}

// insertOrUpdateLiveAndDynamic 插入或更新数据库
func (bdb *bilibilipushdb) insertOrUpdateLiveAndDynamic(bpMap map[string]any) (err error) {
	db := (*gorm.DB)(bdb)
	bp := bilibilipush{}
	data, err := json.Marshal(&bpMap)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &bp)
	if err != nil {
		return
	}
	if err = db.Where("bilibili_uid = ? and group_id = ? and bot_id = ?", bp.BilibiliUID, bp.GroupID).First(&bp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = db.Model(&bilibilipush{}).Create(&bp).Error
		}
	} else {
		err = db.Model(&bilibilipush{}).Where("bilibili_uid = ? and group_id = ? and bot_id = ?", bp.BilibiliUID, bp.GroupID).Updates(bpMap).Error
	}
	return
}

func (bdb *bilibilipushdb) getAllBuidByLive() (buidList []int64) {
	db := (*gorm.DB)(bdb)
	var bpl []bilibilipush
	db.Model(&bilibilipush{}).Find(&bpl, "live_disable = 0")
	temp := make(map[int64]bool)
	for _, v := range bpl {
		_, ok := temp[v.BilibiliUID]
		if !ok {
			buidList = append(buidList, v.BilibiliUID)
			temp[v.BilibiliUID] = true
		}
	}
	return
}

func (bdb *bilibilipushdb) getAllBuidByDynamic() (buidList []int64) {
	db := (*gorm.DB)(bdb)
	var bpl []bilibilipush
	db.Model(&bilibilipush{}).Find(&bpl, "dynamic_disable = 0")
	temp := make(map[int64]bool)
	for _, v := range bpl {
		_, ok := temp[v.BilibiliUID]
		if !ok {
			buidList = append(buidList, v.BilibiliUID)
			temp[v.BilibiliUID] = true
		}
	}
	return
}

func (bdb *bilibilipushdb) getAllGroupByBuidAndLive(buid int64) (groupList []int64) {
	db := (*gorm.DB)(bdb)
	var bpl []bilibilipush
	db.Model(&bilibilipush{}).Find(&bpl, "bilibili_uid = ? and live_disable = 0", buid)
	for _, v := range bpl {
		groupList = append(groupList, v.GroupID)
	}
	return
}

func (bdb *bilibilipushdb) getAllGroupByBuidAndDynamic(buid int64) (groupList []int64) {
	db := (*gorm.DB)(bdb)
	var bpl []bilibilipush
	err := db.Where("bilibili_uid = ? and dynamic_disable = 0", buid).Find(&bpl).Error
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}
	for _, v := range bpl {
		groupList = append(groupList, v.GroupID)
	}
	return
}

func (bdb *bilibilipushdb) getAllPushByGroup(groupID int64) (bpl []bilibilipush) {
	db := (*gorm.DB)(bdb)
	db.Where("group_id = ? and (live_disable = 0 or dynamic_disable = 0)", groupID).Find(&bpl)
	return
}

func (bdb *bilibilipushdb) getAtAll(groupID int64) (res int) {
	db := (*gorm.DB)(bdb)
	var bpl bilibiliAt
	db.Model(&bilibilipush{}).Find(&bpl, "group_id = ?", groupID)
	res = bpl.AtAll
	return
}

func (bdb *bilibilipushdb) updateAtAll(bp bilibiliAt) (err error) {
	db := (*gorm.DB)(bdb)
	err = db.Save(&bp).Error
	return
}

func (bdb *bilibilipushdb) insertBilibiliUp(buid int64, name string) {
	db := (*gorm.DB)(bdb)
	bu := bilibiliup{
		BilibiliUID: buid,
		Name:        name,
	}
	db.Model(&bilibiliup{}).Create(bu)
}

func (bdb *bilibilipushdb) updateAllUp() {
	db := (*gorm.DB)(bdb)
	var bul []bilibiliup
	db.Model(&bilibiliup{}).Find(&bul)
	for _, v := range bul {
		upMap[v.BilibiliUID] = v.Name
	}
}
