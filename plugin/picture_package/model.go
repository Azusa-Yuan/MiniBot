package ymgal

import (
	"MiniBot/utils/path"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// gdb 得分数据库
var gdb *ymgaldb

// ymgaldb galgame图片数据库
type ymgaldb gorm.DB

var mu sync.RWMutex

// ymgal gal图片储存结构体
type ymgal struct {
	ID                 int64  `gorm:"column:id;primary_key" `
	OtherId            int64  `gorm:"column:other_id;index" `
	Title              string `gorm:"column:title;index" `
	PictureType        string `gorm:"column:picture_type" `
	PictureDescription string `gorm:"column:picture_description;type:varchar(1024)" `
	PictureList        string `gorm:"column:picture_list;type:text" `
}

// TableName ...
func (ymgal) TableName() string {
	return "ymgal"
}

func (gdb *ymgaldb) insertOrUpdateYmgalByID(id int64, title, pictureType, pictureDescription, pictureList string) (err error) {
	db := (*gorm.DB)(gdb)
	y := ymgal{
		OtherId:            id,
		Title:              title,
		PictureType:        pictureType,
		PictureDescription: pictureDescription,
		PictureList:        pictureList,
	}
	old := &ymgal{}
	err = db.Model(&ymgal{}).First(&old, "other_id = ? ", id).Error
	if err == nil {
		y.ID = old.ID
	}

	db.Save(&y)
	return
}

func (gdb *ymgaldb) insertOrUpdateLocalPic(title, pictureType, pictureDescription, pictureList string) (err error) {
	db := (*gorm.DB)(gdb)
	y := ymgal{
		Title:              title,
		PictureType:        pictureType,
		PictureDescription: pictureDescription,
		PictureList:        pictureList,
	}
	old := &ymgal{}
	err = db.Model(&ymgal{}).First(&old, "title = ? ", title).Error
	if err == nil {
		y.ID = old.ID
	}

	db.Save(&y)
	return
}

func (gdb *ymgaldb) getYmgalByID(id string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	db.Model(&ymgal{}).Where("other_id = ?", id).Take(&y)
	return
}

func (gdb *ymgaldb) randPicByType(pictureType string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	var count int64
	s := db.Model(&ymgal{}).Where("picture_type = ?", pictureType).Count(&count)
	if count == 0 {
		return
	}
	s.Offset(int(rand.Int64N(count))).Take(&y)
	return
}

func (gdb *ymgaldb) randPicByKey(key string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	var count int64
	var s *gorm.DB
	if key != "" {
		s = db.Model(&ymgal{}).Where("title like ? or picture_description like ?", "%"+key+"%", "%"+key+"%").Count(&count)
	} else {
		s = db.Model(&ymgal{}).Count(&count)
	}
	if count == 0 {
		return
	}
	s.Offset(int(rand.Int64N(count))).Take(&y)
	return
}

func (gdb *ymgaldb) randPicBytypeAndKey(pictureType string, key string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	var count int64
	s := db.Model(&ymgal{}).Where("picture_type = ? and (picture_description like ? or title like ?) ", pictureType, "%"+key+"%", "%"+key+"%").Count(&count)
	if count == 0 {
		return
	}
	s.Offset(int(rand.Int64N(count))).Take(&y)
	return
}

func (gdb *ymgaldb) getRandPic(pictureType, key string) (y ymgal) {
	if pictureType == "" {
		return gdb.randPicByKey(key)
	} else if key == "" {
		return gdb.randPicByType(pictureType)
	}
	return gdb.randPicBytypeAndKey(pictureType, key)
}

func (gdb *ymgaldb) getYmgalByKey(pictureType, key string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	var count int64
	var s *gorm.DB
	if key != "" {
		s = db.Model(&ymgal{}).Where("picture_type = ? and (picture_description like ? or title like ?) ", pictureType, "%"+key+"%", "%"+key+"%").Count(&count)
	} else {
		s = db.Model(&ymgal{}).Where("picture_type = ? and (picture_description like ? or title like ?) ", pictureType, "%"+key+"%", "%"+key+"%").Count(&count)
	}
	if count == 0 {
		return
	}
	s.Offset(int(rand.Int64N(count))).Take(&y)
	return
}

const (
	webURL       = "https://www.ymgal.games"
	cgType       = "cg"
	emoticonType = "emoji"
	webPicURL    = webURL + "/co/picset/"
	reNumber     = `\d+`
)

var (
	cgURL                = webURL + "/search?type=picset&sort=default&category=" + url.QueryEscape(cgType) + "&page="
	emoticonURL          = webURL + "/search?type=picset&sort=default&category=" + url.QueryEscape(emoticonType) + "&page="
	commonPageNumberExpr = "//*[@id='pager-box']/div/a[@class='icon item pager-next']/preceding-sibling::a[1]/text()"
	cgIDList             []string
	emoticonIDList       []string
	dataPath             = path.GetPluginDataPath()
)

func initPageNumber() (maxCgPageNumber, maxEmoticonPageNumber int, err error) {
	doc, err := htmlquery.LoadURL(cgURL + "1")
	if err != nil {
		return
	}
	maxCgPageNumber, err = strconv.Atoi(htmlquery.FindOne(doc, commonPageNumberExpr).Data)
	if err != nil {
		return
	}
	doc, err = htmlquery.LoadURL(emoticonURL + "1")
	if err != nil {
		return
	}
	maxEmoticonPageNumber, err = strconv.Atoi(htmlquery.FindOne(doc, commonPageNumberExpr).Data)
	if err != nil {
		return
	}
	return
}

func getPicID(pageNumber int, pictureType string) error {
	var picURL string
	if pictureType == cgType {
		picURL = cgURL + strconv.Itoa(pageNumber)
	} else if pictureType == emoticonType {
		picURL = emoticonURL + strconv.Itoa(pageNumber)
	}
	doc, err := htmlquery.LoadURL(picURL)
	if err != nil {
		return err
	}
	list := htmlquery.Find(doc, "//*[@id='picset-result-list']/ul/div/div[1]/a")
	for i := 0; i < len(list); i++ {
		re := regexp.MustCompile(reNumber)
		picID := re.FindString(list[i].Attr[0].Val)
		if pictureType == cgType {
			cgIDList = append(cgIDList, picID)
		} else if pictureType == emoticonType {
			emoticonIDList = append(emoticonIDList, picID)
		}
	}
	return nil
}

func updateLocalPic() error {
	titleDirs, err := os.ReadDir(dataPath)
	if err != nil {
		return err
	}
	for _, titleDir := range titleDirs {
		if !titleDir.IsDir() {
			continue
		}
		titleSuffix := titleDir.Name()
		typePath := filepath.Join(dataPath, titleSuffix)
		typeDirs, err := os.ReadDir(typePath)
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
			continue
		}
		for _, typeDir := range typeDirs {
			if !typeDir.IsDir() {
				continue
			}
			picType := typeDir.Name()
			title := picType + "-" + titleSuffix
			picsPath := filepath.Join(typePath, picType)
			picList, err := os.ReadDir(picsPath)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
				continue
			}
			picListStr := ""
			picListDesc := ""

			for _, pic := range picList {
				picPath := filepath.Join(picsPath, pic.Name())
				if strings.HasSuffix(picPath, "txt") {
					descBytes, err := os.ReadFile(picPath)
					if err != nil {
						log.Error().Str("name", pluginName).Err(err).Msg("")
						continue
					}
					picListDesc += string(descBytes)
					continue
				}
				if picListStr != "" {
					picListStr += ","
				}
				picListStr += "file://" + picPath
			}
			err = gdb.insertOrUpdateLocalPic(title, picType, picListDesc, picListStr)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
			}
		}
	}
	return nil
}

func updatePic() error {
	maxCgPageNumber, maxEmoticonPageNumber, err := initPageNumber()
	if err != nil {
		return err
	}
	for i := 1; i <= maxCgPageNumber; i++ {
		err = getPicID(i, cgType)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
	}
	for i := 1; i <= maxEmoticonPageNumber; i++ {
		err = getPicID(i, emoticonType)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
	}
	for i := len(cgIDList) - 1; i >= 0; i-- {
		mu.RLock()
		y := gdb.getYmgalByID(cgIDList[i])
		mu.RUnlock()
		if y.PictureList == "" {
			mu.Lock()
			err = storeCgPic(cgIDList[i])
			mu.Unlock()
			if err != nil {
				return err
			}
		}
		time.Sleep(time.Millisecond * 100)
	}

	for i := len(emoticonIDList) - 1; i >= 0; i-- {
		mu.RLock()
		y := gdb.getYmgalByID(emoticonIDList[i])
		mu.RUnlock()
		if y.PictureList == "" {
			mu.Lock()
			err = storeEmoticonPic(emoticonIDList[i])
			mu.Unlock()
			if err != nil {
				return err
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func storeCgPic(picIDStr string) (err error) {
	picID, err := strconv.ParseInt(picIDStr, 10, 64)
	if err != nil {
		return
	}
	pictureType := cgType
	doc, err := htmlquery.LoadURL(webPicURL + picIDStr)
	if err != nil {
		return
	}
	title := htmlquery.FindOne(doc, "//meta[@name='name']").Attr[1].Val
	pictureDescription := htmlquery.FindOne(doc, "//meta[@name='description']").Attr[1].Val
	pictureNumberStr := htmlquery.FindOne(doc, "//div[@class='meta-info']/div[@class='meta-right']/span[2]/text()").Data
	re := regexp.MustCompile(reNumber)
	pictureNumber, err := strconv.Atoi(re.FindString(pictureNumberStr))
	if err != nil {
		return
	}
	pictureList := ""
	for i := 1; i <= pictureNumber; i++ {
		picURL := htmlquery.FindOne(doc, fmt.Sprintf("//*[@id='main-picset-warp']/div/div[2]/div/div[@class='swiper-wrapper']/div[%d]", i)).Attr[1].Val
		if i == 1 {
			pictureList += picURL
		} else {
			pictureList += "," + picURL
		}
	}
	err = gdb.insertOrUpdateYmgalByID(picID, title, pictureType, pictureDescription, pictureList)
	return
}

func storeEmoticonPic(picIDStr string) error {
	picID, err := strconv.ParseInt(picIDStr, 10, 64)
	if err != nil {
		return err
	}
	pictureType := emoticonType
	doc, err := htmlquery.LoadURL(webPicURL + picIDStr)
	if err != nil {
		return err
	}
	title := htmlquery.FindOne(doc, "//meta[@name='name']").Attr[1].Val
	pictureDescription := htmlquery.FindOne(doc, "//meta[@name='description']").Attr[1].Val
	if !(strings.Contains(title, "表情包") || strings.Contains(pictureDescription, "表情包")) {
		return nil
	}
	pictureNumberStr := htmlquery.FindOne(doc, "//div[@class='meta-info']/div[@class='meta-right']/span[2]/text()").Data
	re := regexp.MustCompile(reNumber)
	pictureNumber, err := strconv.Atoi(re.FindString(pictureNumberStr))
	if err != nil {
		return err
	}
	pictureList := ""
	for i := 1; i <= pictureNumber; i++ {
		picURL := htmlquery.FindOne(doc, fmt.Sprintf("//*[@id='main-picset-warp']/div/div[@class='stream-list']/div[%d]/img", i)).Attr[1].Val
		if i == 1 {
			pictureList += picURL
		} else {
			pictureList += "," + picURL
		}
	}
	return gdb.insertOrUpdateYmgalByID(picID, title, pictureType, pictureDescription, pictureList)
}
