package picturepackage

import (
	"MiniBot/utils/path"
	"fmt"
	"io/fs"
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

// picturePackage gal图片储存结构体
type picturePackage struct {
	ID                 int64  `gorm:"column:id;primary_key" `
	OtherId            int64  `gorm:"column:other_id;index" `
	Title              string `gorm:"column:title;index" `
	PictureType        string `gorm:"column:picture_type" `
	PictureDescription string `gorm:"column:picture_description;type:varchar(1024)" `
	PictureList        string `gorm:"column:picture_list;type:text" `
}

func (gdb *ymgaldb) insertOrUpdateYmgalByID(id int64, title, pictureType, pictureDescription, pictureList string) (err error) {
	db := (*gorm.DB)(gdb)
	y := picturePackage{
		OtherId:            id,
		Title:              title,
		PictureType:        pictureType,
		PictureDescription: pictureDescription,
		PictureList:        pictureList,
	}
	old := &picturePackage{}
	err = db.Model(&picturePackage{}).First(&old, "other_id = ? ", id).Error
	if err == nil {
		y.ID = old.ID
	}

	err = db.Save(&y).Error
	return
}

func (gdb *ymgaldb) insertOrUpdateLocalPic(title, pictureType, pictureDescription, pictureList string) (err error) {
	db := (*gorm.DB)(gdb)
	y := picturePackage{
		Title:              title,
		PictureType:        pictureType,
		PictureDescription: pictureDescription,
		PictureList:        pictureList,
	}
	old := &picturePackage{}
	err = db.Model(&picturePackage{}).First(&old, "title = ? ", title).Error
	if err == nil {
		y.ID = old.ID
	}

	err = db.Save(&y).Error
	return
}

func (gdb *ymgaldb) getYmgalByID(id string) (y picturePackage) {
	db := (*gorm.DB)(gdb)
	db.Model(&picturePackage{}).Where("other_id = ?", id).Take(&y)
	return
}

func (gdb *ymgaldb) randPicByType(pictureType string) (y picturePackage) {
	db := (*gorm.DB)(gdb)
	var count int64
	s := db.Model(&picturePackage{}).Where("picture_type = ?", pictureType).Count(&count)
	if count == 0 {
		return
	}
	s.Offset(int(rand.Int64N(count))).Take(&y)
	return
}

func (gdb *ymgaldb) randPicByKey(key string) (y picturePackage) {
	db := (*gorm.DB)(gdb)
	var count int64
	var s *gorm.DB
	if key != "" {
		s = db.Model(&picturePackage{}).Where("title like ? or picture_description like ?", "%"+key+"%", "%"+key+"%").Count(&count)
	} else {
		s = db.Model(&picturePackage{}).Count(&count)
	}
	if count == 0 {
		return
	}
	s.Offset(int(rand.Int64N(count))).Take(&y)
	return
}

func (gdb *ymgaldb) randPicBytypeAndKey(pictureType string, key string) (y picturePackage) {
	db := (*gorm.DB)(gdb)
	var count int64
	s := db.Model(&picturePackage{}).Where("picture_type = ? and (picture_description like ? or title like ?) ", pictureType, "%"+key+"%", "%"+key+"%").Count(&count)
	if count == 0 {
		return
	}
	s.Offset(int(rand.Int64N(count))).Take(&y)
	return
}

func (gdb *ymgaldb) getRandPic(pictureType, key string) (y picturePackage) {
	if pictureType == "" {
		return gdb.randPicByKey(key)
	} else if key == "" {
		return gdb.randPicByType(pictureType)
	}
	return gdb.randPicBytypeAndKey(pictureType, key)
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

func deepWalk(fsys fs.FS, path, desc string) error {
	if strings.Contains(path, "18") {
		return nil
	}

	dirEntries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return err
	}

	picListStr := ""
	dirList := []string{}
	picType := typeMap[(strings.Split(path, "/")[0])]
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		curPath := filepath.Join(path, name)

		if dirEntry.IsDir() {
			dirList = append(dirList, curPath)
			continue
		}

		if strings.HasSuffix(name, "txt") {
			descBytes, err := fs.ReadFile(fsys, curPath)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
				continue
			}
			desc += string(descBytes)
			continue
		}

		if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") ||
			strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".gif") || strings.HasSuffix(name, ".webp") {
			if picListStr != "" {
				picListStr += ","
			}
			picListStr += "file://" + filepath.Join(dataPath, curPath)
		}
	}

	if picListStr != "" {
		err = gdb.insertOrUpdateLocalPic(path, picType, desc, picListStr)
	}
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
	}

	for _, dirPath := range dirList {
		err = deepWalk(fsys, dirPath, desc)
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
		}
	}

	return nil
}

func updateLocalPic() error {
	root := os.DirFS(dataPath)
	deepWalk(root, ".", "")
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
			err = storeYmgalPic(cgIDList[i], cgType)
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

func storeYmgalPic(picIDStr, pictureType string) (err error) {
	picID, err := strconv.ParseInt(picIDStr, 10, 64)
	if err != nil {
		return
	}
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
		htmlNode := htmlquery.FindOne(doc, fmt.Sprintf("//*[@id='main-picset-warp']/div/div[2]/div/div[@class='swiper-wrapper']/div[%d]", i))
		if htmlNode == nil || len(htmlNode.Attr) < 2 {
			log.Info().Str("name", pluginName).Msg("can not get " + webPicURL + picIDStr)
			continue
		}
		picURL := htmlNode.Attr[1].Val
		if pictureList == "" {
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
		htmlNode := htmlquery.FindOne(doc, fmt.Sprintf("//*[@id='main-picset-warp']/div/div[@class='stream-list']/div[%d]/img", i))
		if htmlNode == nil || len(htmlNode.Attr) < 2 {
			log.Info().Str("name", pluginName).Msg("can not get " + webPicURL + picIDStr)
			continue
		}
		picURL := htmlNode.Attr[1].Val
		if pictureList == "" {
			pictureList += picURL
		} else {
			pictureList += "," + picURL
		}
	}
	return gdb.insertOrUpdateYmgalByID(picID, title, pictureType, pictureDescription, pictureList)
}
