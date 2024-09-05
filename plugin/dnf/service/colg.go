package service

import (
	"MiniBot/utils/path"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/xrash/smetrics"
)

var (
	urlList = []string{
		"https://bbs.colg.cn/home.php?mod=space&uid=4120473",
		"https://bbs.colg.cn/home.php?mod=space&uid=80727",
	}
	authorList = []string{"白猫之惩戒", "魔法少女QB"}
	keyWords   = []string{"韩服", "爆料", "国服", "前瞻", "韩测"}
	limit      = 6
	// 缓存不同作者的同一个消息
	preHead     []string
	defaultUser *user
)

type user struct {
	Group []string `json:"group"`
	QQ    []string `json:"qq"`
}

func GetColgUser() (*user, error) {

	if defaultUser != nil {
		return defaultUser, nil
	}

	userPath := filepath.Join(path.GetPluginDataPath(), "user.json")

	users := user{
		[]string{},
		[]string{},
	}

	// 检查文件是否存在
	if _, err := os.Stat(userPath); err == nil {
		// 读取文件内容
		data, err := os.ReadFile(userPath)
		if err != nil {
			logrus.Errorln("读取文件错误:", err)
			return nil, err
		}

		// 解析 JSON 数据
		err = json.Unmarshal(data, &users)
		if err != nil {
			logrus.Errorln("解析 JSON 数据错误:", err)
			return nil, err
		}
	}

	defaultUser = &users
	return &users, nil
}

func (user) saveBinds(user *user, userPath string) {
	// 序列化 JSON 数据
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		logrus.Infoln("序列化 JSON 数据错误:", err)
		return
	}

	// 写入文件
	err = os.WriteFile(userPath, data, 0644)
	if err != nil {
		logrus.Infoln("写入文件错误:", err)
		return
	}
}

func fetchContent(url string) (string, string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}

	var context strings.Builder
	var head string

	doc.Find("#thread_content li").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i < limit {
			context.WriteString(s.Text() + "\r\n")
			link, _ := s.Find("a").Attr("href")
			context.WriteString("https://bbs.colg.cn/" + link + "\r\n")
			if i == 0 {
				head = s.Text()
			}
			return true
		}
		return false // Stop iteration
	})

	return context.String(), head, nil
}

func ColgNews() (string, error) {
	var context strings.Builder
	for _, url := range urlList {
		tmpContext, _, err := fetchContent(url)
		if err != nil {
			return "", err
		}
		context.WriteString(tmpContext)
	}
	return context.String(), nil
}

func (user) GetChange() ([]string, error) {
	var newList []string

	if len(preHead) == 0 {
		for _, url := range urlList {
			_, tmpHead, err := fetchContent(url)
			if err != nil {
				return nil, err
			}
			preHead = append(preHead, tmpHead)
		}
		return nil, nil
	}

	for order, url := range urlList {
		context, head, err := fetchContent(url)
		if err != nil {
			return nil, err
		}
		if head != preHead[order] {
			for _, keyWord := range keyWords {
				if strings.Contains(head, keyWord) && smetrics.Jaro(head, preHead[order]) < 0.8 {
					context = "colg资讯已更新:\r\n" + context + authorList[order]
					newList = append(newList, context)
					break
				}
			}
			preHead[order] = head
		}
	}

	return newList, nil
}
