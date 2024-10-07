package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v65/github"
)

var (
	// 时间间隔 单位小时
	interval = 12
	client   = github.NewClient(nil)
)

func GetCommit() {
	author := "LagrangeDev"
	repo := "go-cqhttp"
	commits, resp, err := client.Repositories.ListCommits(context.Background(), "LagrangeDev", "go-cqhttp", &github.CommitsListOptions{})
	if err != nil {
		return
	}
	commit := commits[0]
	timeStamp := commit.Commit.Author.GetDate().Add(time.Duration(interval) * time.Hour)
	if time.Now().Before(timeStamp) {
		fmt.Println(commit.Commit.GetAuthor().GetName() + "在仓库" + author + "/" +
			repo + "进行了commit " + commit.Commit.GetMessage() + "\n" + commit.GetHTMLURL())
		fmt.Println(resp)
	}

}
