package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	preHead []string
)

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

func GetColgChange() ([]string, error) {
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
