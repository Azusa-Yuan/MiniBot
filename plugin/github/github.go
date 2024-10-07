package github

import (
	"MiniBot/service/book"
	"MiniBot/utils"
	database "MiniBot/utils/db"
	"MiniBot/utils/schedule"
	zero "ZeroBot"
	"ZeroBot/message"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v65/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
)

var pluginName = "github"
var db = database.GetDefalutDB()

// qps要求不高  直接大锁就好了
var githubLock = sync.RWMutex{}

func init() {
	metadata := &zero.Metadata{
		Name: pluginName,
		Help: fmt.Sprintf(`订阅github commit动态, 当前查询间隔为%d小时
指令如下:
订阅github 作者/参数    如:订阅github Azusa-Yuan/MiniBot
取消订阅github 序号     如:取消订阅github 0
查看github订阅`, interval),
		Level: 2,
	}
	engine := zero.NewTemplate(metadata)

	engine.OnPrefix("订阅github").SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			params := strings.Split(param, "/")
			if len(params) != 2 {
				ctx.SendChain(message.At(ctx.Event.SelfID), message.Text("订阅githu参数错误"))
				return
			}
			_, resp, err := client.Repositories.ListCommits(context.Background(), params[0], params[1], nil)
			if err != nil {
				ctx.SendError(err)
				return
			}
			if resp.StatusCode != 200 {
				ctx.SendChain(message.At(ctx.Event.SelfID), message.Text(resp.Status))
				return
			}

			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			bid := ctx.Event.SelfID
			if gid != 0 {
				uid = 0
			}
			err = CreateOrUpdate(book.Book{
				UserID:  uid,
				BotID:   bid,
				GroupID: gid,
				Service: pluginName,
			}, param)
			if err != nil {
				ctx.SendError(err)
				return
			}
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("github订阅成功"))
		},
	)

	schedule.Cron.AddFunc("0 * * * *", sendChange)
}

func CreateOrUpdate(userInfo book.Book, param string) error {
	githubLock.Lock()
	defer githubLock.Unlock()
	bookInfo, err := book.GetUserBookInfo(&userInfo)
	if err != nil {
		return err
	}

	githubRepos := []string{}
	if bookInfo.Value != "" {
		err := json.Unmarshal(utils.StringToBytes(bookInfo.Value), &githubRepos)
		if err != nil {
			return err
		}
		for _, repo := range githubRepos {
			if repo == param {
				return nil
			}
		}
	}
	githubRepos = append(githubRepos, param)
	infosBytes, err := json.MarshalIndent(githubRepos, "", " ")
	if err != nil {
		return err
	}
	bookInfo.Value = utils.BytesToString(infosBytes)
	return db.Save(&bookInfo).Error
}

func sendChange() {
	userInfos, err := book.GetBookInfos(pluginName)
	if err != nil {
		log.Error().Str("name", pluginName).Err(err).Msg("")
		return
	}

	// 去重
	githubMap := map[string]struct{}{}
	for _, info := range userInfos {
		githubRepos := []string{}
		err := json.Unmarshal(utils.StringToBytes(info.Value), &githubRepos)
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
			continue
		}
		for _, repo := range githubRepos {
			githubMap[repo] = struct{}{}
		}
	}

	allResult := map[string][]*github.RepositoryCommit{}
	repos := maps.Keys(githubMap)
	for _, repo := range repos {
		params := strings.Split(repo, "/")
		commits, resp, err := client.Repositories.ListCommits(context.Background(), params[0], params[1], nil)
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
			continue
		}
		if resp.StatusCode != 200 {
			data, err := io.ReadAll(resp.Response.Body)
			if err != nil {
				log.Error().Str("name", pluginName).Err(err).Msg("")
				continue
			}
			log.Error().Str("name", pluginName).Msg(fmt.Sprint(resp.Status, data))
			return
		}
		allResult[repo] = commits
		time.Sleep(1 * time.Second)
	}

	timeStamp := time.Now().Add(-time.Duration(interval) * time.Hour).Add(-5 * time.Minute)
	for _, info := range userInfos {
		githubRepos := []string{}
		err := json.Unmarshal(utils.StringToBytes(info.Value), &githubRepos)
		if err != nil {
			log.Error().Str("name", pluginName).Err(err).Msg("")
			continue
		}
		for _, repo := range githubRepos {
			commits := allResult[repo]
			for _, commit := range commits {
				bot, err := zero.GetBot(info.BotID)
				if err != nil {
					log.Error().Str("name", pluginName).Err(err).Msg("")
					continue
				}
				if commit.Commit.Author.GetDate().After(timeStamp) {
					commitInfo := fmt.Sprintf(commit.Commit.GetAuthor().GetName() + "在仓库" + repo + "进行了commit " +
						commit.Commit.GetMessage() + "\n" + commit.GetHTMLURL())

					if info.GroupID != 0 {
						bot.SendGroupMessage(info.GroupID, commitInfo)
					} else {
						bot.SendPrivateMessage(info.UserID, commitInfo)
					}
				} else {
					break
				}
			}
		}
	}
}
